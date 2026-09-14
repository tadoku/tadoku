// Package testketo starts isolated, in-memory Keto servers for tests.
package testketo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

const (
	startupTimeout  = 10 * time.Second
	requestTimeout  = 2 * time.Second
	shutdownTimeout = 3 * time.Second
)

// Fixture owns one Keto process and its in-memory relationship store.
type Fixture struct {
	command  *exec.Cmd
	wait     chan struct{}
	waitErr  error
	dir      string
	logPath  string
	readURL  string
	writeURL string

	closeOnce sync.Once
	closeErr  error
}

// New starts Keto with the repository's production OPL namespaces.
func New(ctx context.Context) (result *Fixture, resultErr error) {
	keto, err := runfiles.Rlocation("keto_v25_4_0/keto")
	if err != nil {
		return nil, fmt.Errorf("locate Keto executable: %w", err)
	}
	namespaces, err := runfiles.Rlocation("tadoku/infra/dev/ory/namespaces.keto.ts")
	if err != nil {
		return nil, fmt.Errorf("locate Keto namespaces: %w", err)
	}

	dir, err := os.MkdirTemp("", "tadoku-testketo-")
	if err != nil {
		return nil, fmt.Errorf("create Keto directory: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			resultErr = errors.Join(resultErr, os.RemoveAll(dir))
		}
	}()

	readAddress := filepath.Join(dir, "read-address")
	writeAddress := filepath.Join(dir, "write-address")
	configPath := filepath.Join(dir, "keto.json")
	logPath := filepath.Join(dir, "keto.log")
	config := map[string]any{
		"dsn": "memory",
		"namespaces": map[string]string{
			"location": fileURL(namespaces),
		},
		"serve": map[string]any{
			"read": map[string]any{
				"host":              "127.0.0.1",
				"port":              0,
				"write_listen_file": fileURL(readAddress),
			},
			"write": map[string]any{
				"host":              "127.0.0.1",
				"port":              0,
				"write_listen_file": fileURL(writeAddress),
			},
			"metrics": map[string]any{
				"host": "127.0.0.1",
				"port": 0,
			},
			"opl": map[string]any{
				"host": "127.0.0.1",
				"port": 0,
			},
		},
	}
	encodedConfig, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("encode Keto config: %w", err)
	}
	if err := os.WriteFile(configPath, encodedConfig, 0o600); err != nil {
		return nil, fmt.Errorf("write Keto config: %w", err)
	}
	logs, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open Keto log: %w", err)
	}

	command := exec.Command(keto, "serve", "--sqa-opt-out", "--config", configPath)
	command.Env = []string{
		"HOME=" + dir,
		"TMPDIR=" + dir,
	}
	command.Stdout = logs
	command.Stderr = logs
	if err := command.Start(); err != nil {
		_ = logs.Close()
		return nil, fmt.Errorf("start Keto: %w", err)
	}

	fixture := &Fixture{
		command: command,
		wait:    make(chan struct{}),
		dir:     dir,
		logPath: logPath,
	}
	go func() {
		fixture.waitErr = command.Wait()
		_ = logs.Close()
		close(fixture.wait)
	}()

	startupContext, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	if err := fixture.waitUntilReady(startupContext, readAddress, writeAddress); err != nil {
		closeErr := fixture.Close()
		return nil, errors.Join(err, closeErr)
	}

	cleanup = false
	return fixture, nil
}

func (fixture *Fixture) waitUntilReady(ctx context.Context, readAddress, writeAddress string) error {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-fixture.wait:
			return fixture.unexpectedExit("during startup")
		case <-ctx.Done():
			return fmt.Errorf("wait for Keto readiness: %w%s", ctx.Err(), fixture.logs())
		case <-ticker.C:
			readURL, readErr := listenURL(readAddress)
			writeURL, writeErr := listenURL(writeAddress)
			if readErr != nil || writeErr != nil {
				continue
			}

			request, err := http.NewRequestWithContext(ctx, http.MethodGet, readURL+"/health/ready", nil)
			if err != nil {
				return err
			}
			response, err := (&http.Client{Timeout: requestTimeout}).Do(request)
			if err != nil {
				continue
			}
			_ = response.Body.Close()
			if response.StatusCode != http.StatusOK {
				continue
			}

			fixture.readURL = readURL
			fixture.writeURL = writeURL
			return nil
		}
	}
}

// Reset deletes every relationship in every namespace owned by the fixture,
// then loads any provided JSON arrays of relationship tuples.
func (fixture *Fixture) Reset(ctx context.Context, seedFiles ...string) error {
	namespaces, err := fixture.namespaces(ctx)
	if err != nil {
		return err
	}
	for _, namespace := range namespaces {
		endpoint := fixture.writeURL + "/admin/relation-tuples?namespace=" + url.QueryEscape(namespace)
		if err := fixture.do(ctx, http.MethodDelete, endpoint, nil); err != nil {
			return fmt.Errorf("clear Keto namespace %q: %w", namespace, err)
		}
	}

	for _, seedFile := range seedFiles {
		contents, err := os.ReadFile(seedFile)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read Keto seed %s: %w", seedFile, err)
		}

		var relationships []json.RawMessage
		if err := json.Unmarshal(contents, &relationships); err != nil {
			return fmt.Errorf("decode Keto seed %s: %w", seedFile, err)
		}
		for index, relationship := range relationships {
			if err := fixture.do(ctx, http.MethodPut, fixture.writeURL+"/admin/relation-tuples", relationship); err != nil {
				return fmt.Errorf("seed Keto relationship %s[%d]: %w", seedFile, index, err)
			}
		}
	}

	return nil
}

func (fixture *Fixture) namespaces(ctx context.Context) ([]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fixture.readURL+"/namespaces", nil)
	if err != nil {
		return nil, err
	}
	response, err := (&http.Client{Timeout: requestTimeout}).Do(request)
	if err != nil {
		return nil, fmt.Errorf("discover Keto namespaces: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("discover Keto namespaces: status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		Namespaces []struct {
			Name string `json:"name"`
		} `json:"namespaces"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode Keto namespaces: %w", err)
	}
	namespaces := make([]string, 0, len(result.Namespaces))
	for _, namespace := range result.Namespaces {
		if namespace.Name == "" {
			return nil, fmt.Errorf("Keto returned an unnamed namespace")
		}
		namespaces = append(namespaces, namespace.Name)
	}
	return namespaces, nil
}

func (fixture *Fixture) do(ctx context.Context, method, endpoint string, body []byte) error {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := (&http.Client{Timeout: requestTimeout}).Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("%s %s: status %d: %s", method, endpoint, response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

// Close stops and reaps Keto. It reports a process that exited before cleanup.
func (fixture *Fixture) Close() error {
	fixture.closeOnce.Do(func() {
		defer func() {
			fixture.closeErr = errors.Join(fixture.closeErr, os.RemoveAll(fixture.dir))
		}()

		select {
		case <-fixture.wait:
			fixture.closeErr = fixture.unexpectedExit("before close")
			return
		default:
		}

		signalErr := fixture.command.Process.Signal(os.Interrupt)
		timer := time.NewTimer(shutdownTimeout)
		defer timer.Stop()
		select {
		case <-fixture.wait:
			if signalErr != nil || fixture.waitErr != nil {
				fixture.closeErr = fmt.Errorf("stop Keto: %v%s", errors.Join(signalErr, fixture.waitErr), fixture.logs())
			}
			return
		case <-timer.C:
		}

		killErr := fixture.command.Process.Kill()
		<-fixture.wait
		fixture.closeErr = errors.Join(
			fmt.Errorf("Keto did not stop within %s%s", shutdownTimeout, fixture.logs()),
			killErr,
			fixture.waitErr,
		)
	})
	return fixture.closeErr
}

// ReadURL returns the read API of the owned Keto process.
func (fixture *Fixture) ReadURL() string { return fixture.readURL }

// WriteURL returns the write API of the owned Keto process.
func (fixture *Fixture) WriteURL() string { return fixture.writeURL }

func (fixture *Fixture) unexpectedExit(when string) error {
	if fixture.waitErr == nil {
		return fmt.Errorf("Keto exited successfully %s%s", when, fixture.logs())
	}
	return fmt.Errorf("Keto exited %s: %w%s", when, fixture.waitErr, fixture.logs())
}

func (fixture *Fixture) logs() string {
	contents, err := os.ReadFile(fixture.logPath)
	if err != nil || len(contents) == 0 {
		return ""
	}
	return "\nKeto logs:\n" + string(contents)
}

func fileURL(path string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String()
}

func listenURL(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	address := strings.TrimSpace(string(contents))
	if address == "" {
		return "", fmt.Errorf("empty listen address")
	}
	return "http://" + address, nil
}
