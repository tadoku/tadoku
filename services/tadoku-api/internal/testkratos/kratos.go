// Package testkratos owns disposable Kratos processes with RAM-backed SQLite.
// It is test-only and never accepts an external provider or database address.
package testkratos

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	_ "github.com/mattn/go-sqlite3"
	kratosapi "github.com/ory/kratos-client-go"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
)

const (
	startupTimeout  = 30 * time.Second
	requestTimeout  = 2 * time.Second
	shutdownTimeout = 3 * time.Second
	adminURL        = "http://kratos.test"
)

// Fixture is shared only by sequential tests. Its client survives Reset.
// A failed reset makes the fixture unusable until Close.
type Fixture struct {
	dir        string
	binary     string
	configPath string
	seed       []byte

	client    *kratosapi.APIClient
	http      *http.Client
	transport *http.Transport

	command *exec.Cmd
	wait    chan struct{}
	waitErr error
	failed  error
	closed  bool

	closeOnce sync.Once
	closeErr  error
}

// New migrates a fresh RAM-backed database, starts Kratos and applies seedFile
// once. Seeds are trusted repository SQL for the pinned provider's schema.
func New(ctx context.Context, seedFile string) (_ *Fixture, resultErr error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	seed, err := os.ReadFile(seedFile)
	if err != nil {
		return nil, fmt.Errorf("read Kratos seed: %w", err)
	}
	binary, err := runfiles.Rlocation("kratos_v26_2_0/kratos")
	if err != nil {
		return nil, fmt.Errorf("locate Kratos executable: %w", err)
	}
	schema, err := runfiles.Rlocation("tadoku/infra/dev/ory/identity.default.schema.json")
	if err != nil {
		return nil, fmt.Errorf("locate Kratos identity schema: %w", err)
	}

	var filesystem syscall.Statfs_t
	if err := syscall.Statfs("/dev/shm", &filesystem); err != nil {
		return nil, fmt.Errorf("Kratos fixture requires writable /dev/shm: %w", err)
	}
	if filesystem.Type != 0x01021994 { // Linux TMPFS_MAGIC.
		return nil, errors.New("Kratos fixture requires tmpfs at /dev/shm")
	}
	dir, err := os.MkdirTemp("/dev/shm", fmt.Sprintf("tadoku-testkratos-%d-", os.Getpid()))
	if err != nil {
		return nil, fmt.Errorf("create Kratos RAM-backed directory: %w", err)
	}
	fixture := &Fixture{
		dir:    dir,
		binary: binary,
		seed:   seed,
	}
	defer func() {
		if resultErr != nil {
			resultErr = errors.Join(resultErr, fixture.Close())
		}
	}()

	dialer := &net.Dialer{Timeout: requestTimeout}
	fixture.transport = &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", filepath.Join(dir, "admin.sock"))
		},
		MaxConnsPerHost:       4,
		MaxIdleConns:          4,
		MaxIdleConnsPerHost:   4,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: requestTimeout,
	}
	fixture.http = &http.Client{
		Transport: fixture.transport,
		Timeout:   requestTimeout,
	}
	fixture.client = kratosclient.NewAPIClient(adminURL, kratosclient.WithHTTPClient(fixture.http))

	if err := fixture.writeConfig(schema); err != nil {
		return nil, err
	}
	if err := fixture.initialize(ctx); err != nil {
		return nil, err
	}
	return fixture, nil
}

// Client returns the raw SDK, using the fixture's bounded, private transport.
func (fixture *Fixture) Client() *kratosapi.APIClient { return fixture.client }

// CursorClient returns the production cursor-pagination client over the same
// bounded private transport as Client.
func (fixture *Fixture) CursorClient() *kratosclient.Client {
	return kratosclient.NewClient(adminURL, kratosclient.WithHTTPClient(fixture.http))
}

// Err prevents another scenario from using a closed, failed or exited fixture.
func (fixture *Fixture) Err() error {
	if fixture.failed != nil {
		return fixture.failed
	}
	if fixture.closed {
		return errors.New("Kratos fixture is closed")
	}
	if fixture.wait != nil {
		select {
		case <-fixture.wait:
			return fixture.unexpectedExit()
		default:
		}
	}
	return nil
}

// Reset replaces all provider state with the original seed. Call it only after
// a test explicitly marked as mutating Kratos; ordinary cases share the seed.
func (fixture *Fixture) Reset(ctx context.Context) (err error) {
	if err := fixture.Err(); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			fixture.failed = fmt.Errorf("Kratos reset failed: %w", err)
		}
	}()
	fixture.transport.CloseIdleConnections()
	if err := fixture.stop(); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(fixture.dir, "database")); err != nil {
		return fmt.Errorf("remove Kratos database: %w", err)
	}
	return fixture.initialize(ctx)
}

func (fixture *Fixture) writeConfig(schema string) error {
	// Keep watched configuration separate from sockets, databases and logs.
	directory := filepath.Join(fixture.dir, "config")
	if err := os.Mkdir(directory, 0o700); err != nil {
		return err
	}
	fixture.configPath = filepath.Join(directory, "kratos.json")
	config := map[string]any{
		"dsn":     "sqlite://file:" + fixture.databasePath() + "?_fk=true",
		"version": "v26.2.0",
		"identity": map[string]any{
			"default_schema_id": "user",
			"schemas": []map[string]string{{
				"id":  "user",
				"url": (&url.URL{Scheme: "file", Path: schema}).String(),
			}},
		},
		"serve": map[string]any{
			"admin": map[string]any{
				"host":     "unix:" + filepath.Join(fixture.dir, "admin.sock"),
				"base_url": adminURL,
				"socket":   map[string]int{"mode": 0o600},
			},
			"public": map[string]any{
				"host":     "unix:" + filepath.Join(fixture.dir, "public.sock"),
				"base_url": "http://kratos-public.test",
				"socket":   map[string]int{"mode": 0o600},
			},
		},
		"selfservice": map[string]string{"default_browser_return_url": "http://kratos-public.test/"},
		"courier":     map[string]any{"smtp": map[string]string{"connection_uri": "smtp://127.0.0.1:1/"}},
		"secrets":     map[string][]string{"default": {"disposable-kratos-fixture-secret-000000"}},
		"log":         map[string]string{"level": "error"},
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(fixture.configPath, data, 0o600)
}

func (fixture *Fixture) databasePath() string {
	return filepath.Join(fixture.dir, "database", "kratos.sqlite")
}

func (fixture *Fixture) initialize(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Dir(fixture.databasePath()), 0o700); err != nil {
		return err
	}

	logs, err := os.OpenFile(filepath.Join(fixture.dir, "kratos.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	migration := exec.CommandContext(ctx, fixture.binary, "migrate", "sql", "--read-from-env", "--yes", "--config", fixture.configPath)
	migration.Env = fixture.environment()
	migration.Stdout = logs
	migration.Stderr = logs
	if err := migration.Run(); err != nil {
		_ = logs.Close()
		return fmt.Errorf("migrate Kratos: %w%s", errors.Join(err, ctx.Err()), fixture.logs())
	}

	command := exec.Command(fixture.binary, "serve", "--dev", "--sqa-opt-out", "--config", fixture.configPath)
	command.Env = fixture.environment()
	command.Stdout = logs
	command.Stderr = logs
	if err := command.Start(); err != nil {
		_ = logs.Close()
		return fmt.Errorf("start Kratos: %w", err)
	}
	fixture.command = command
	fixture.wait = make(chan struct{})
	fixture.waitErr = nil
	go func() {
		fixture.waitErr = command.Wait()
		_ = logs.Close()
		close(fixture.wait)
	}()

	if err := fixture.waitUntilReady(ctx); err != nil {
		return err
	}
	return fixture.seedDatabase(ctx)
}

func (fixture *Fixture) environment() []string {
	return []string{"HOME=" + fixture.dir, "TMPDIR=" + fixture.dir}
}

func (fixture *Fixture) waitUntilReady(ctx context.Context) error {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-fixture.wait:
			return fixture.unexpectedExit()
		case <-ctx.Done():
			return fmt.Errorf("wait for Kratos readiness: %w%s", ctx.Err(), fixture.logs())
		case <-ticker.C:
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, adminURL+"/health/ready", nil)
			if err != nil {
				return err
			}
			response, err := fixture.http.Do(request)
			if err != nil {
				continue
			}
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

func (fixture *Fixture) seedDatabase(ctx context.Context) (err error) {
	db, err := sql.Open("sqlite3", "file:"+fixture.databasePath()+"?_foreign_keys=on&_busy_timeout=1000")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	db.SetMaxOpenConns(1)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var networks int
	if err := tx.QueryRowContext(ctx, "select count(*) from networks").Scan(&networks); err != nil {
		return err
	}
	if networks != 1 {
		return fmt.Errorf("Kratos initialized %d networks, want exactly one", networks)
	}
	if _, err := tx.ExecContext(ctx, string(fixture.seed)); err != nil {
		return fmt.Errorf("seed Kratos identities: %w", err)
	}
	return tx.Commit()
}

func (fixture *Fixture) stop() error {
	if fixture.command == nil {
		return nil
	}
	defer func() { fixture.command = nil }()
	select {
	case <-fixture.wait:
		return fixture.unexpectedExit()
	default:
	}

	signalErr := fixture.command.Process.Signal(os.Interrupt)
	timer := time.NewTimer(shutdownTimeout)
	defer timer.Stop()
	select {
	case <-fixture.wait:
		if err := errors.Join(signalErr, fixture.waitErr); err != nil {
			return fmt.Errorf("stop Kratos: %w%s", err, fixture.logs())
		}
		return nil
	case <-timer.C:
		killErr := fixture.command.Process.Kill()
		<-fixture.wait
		return fmt.Errorf("Kratos exceeded shutdown deadline: %w%s", errors.Join(context.DeadlineExceeded, killErr, fixture.waitErr), fixture.logs())
	}
}

// Close closes connections, stops/reaps the process and removes its owned files.
func (fixture *Fixture) Close() error {
	fixture.closeOnce.Do(func() {
		fixture.closed = true
		if fixture.transport != nil {
			fixture.transport.CloseIdleConnections()
		}
		fixture.closeErr = errors.Join(fixture.stop(), os.RemoveAll(fixture.dir))
	})
	return fixture.closeErr
}

func (fixture *Fixture) unexpectedExit() error {
	return fmt.Errorf("Kratos exited unexpectedly (%v)%s", fixture.waitErr, fixture.logs())
}

func (fixture *Fixture) logs() string {
	file, err := os.Open(filepath.Join(fixture.dir, "kratos.log"))
	if err != nil {
		return ""
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return ""
	}
	if info.Size() > 8192 {
		if _, err := file.Seek(-8192, io.SeekEnd); err != nil {
			return ""
		}
	}
	data, _ := io.ReadAll(io.LimitReader(file, 8192))
	return "\nKratos logs:\n" + string(data)
}
