package valkey

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestOpenRunsRawCommandsAgainstValkey(t *testing.T) {
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	client, err := Open(t.Context(), rawURL, time.Second)
	if err != nil {
		t.Fatalf("open Valkey: %v", err)
	}
	t.Cleanup(client.Close)

	key := "tadoku-api-test:" + randomID(t)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = client.Do(ctx, client.B().Del().Key(key).Build()).Error()
	})

	if err := client.Do(t.Context(), client.B().Set().Key(key).Value("value").Build()).Error(); err != nil {
		t.Fatalf("SET: %v", err)
	}
	value, err := client.Do(t.Context(), client.B().Get().Key(key).Build()).ToString()
	if err != nil || value != "value" {
		t.Fatalf("GET value=%q error=%v", value, err)
	}
	if deleted, err := client.Do(t.Context(), client.B().Del().Key(key).Build()).AsInt64(); err != nil || deleted != 1 {
		t.Fatalf("DEL count=%d error=%v", deleted, err)
	}
	if err := client.Do(t.Context(), client.B().Get().Key(key).Build()).Error(); !valkeygo.IsValkeyNil(err) {
		t.Fatalf("missing GET error=%v", err)
	}
}

func TestOpenDetachesSuccessfulClientFromStartupContext(t *testing.T) {
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	startup, cancel := context.WithCancel(t.Context())
	client, err := Open(startup, rawURL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	cancel()

	ctx, commandCancel := context.WithTimeout(t.Context(), time.Second)
	defer commandCancel()
	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		t.Fatalf("client retained canceled startup context: %v", err)
	}
}

func TestOpenReturnsClientAfterInitialFailureAndReconnects(t *testing.T) {
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	target, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}

	reserved, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := reserved.Addr().String()
	if err := reserved.Close(); err != nil {
		t.Fatal(err)
	}

	startup, cancelStartup := context.WithCancel(t.Context())
	client, initialErr := Open(startup, "redis://"+address, 100*time.Millisecond)
	if client == nil || initialErr == nil {
		t.Fatalf("initial client=%v error=%v", client, initialErr)
	}
	cancelStartup()
	t.Cleanup(client.Close)
	proxy := startTCPProxy(t, address, target.Host)
	t.Cleanup(func() { _ = proxy.Close() })

	key := "tadoku-api-recovery-test:" + randomID(t)
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	for {
		err = client.Do(ctx, client.B().Set().Key(key).Value("recovered").Build()).Error()
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			t.Fatalf("client did not recover: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer func() {
		_ = client.Do(context.Background(), client.B().Del().Key(key).Build()).Error()
	}()

	value, err := client.Do(ctx, client.B().Get().Key(key).Build()).ToString()
	if err != nil || value != "recovered" {
		t.Fatalf("GET after recovery value=%q error=%v", value, err)
	}
}

func TestOpenCancellationClosesStalledHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan net.Conn, 1)
	go func() {
		connection, err := listener.Accept()
		if err == nil {
			accepted <- connection
		}
	}()

	ctx, cancel := context.WithCancel(t.Context())
	result := make(chan error, 1)
	go func() {
		_, err := Open(ctx, "redis://"+listener.Addr().String(), 5*time.Second)
		result <- err
	}()

	connection := <-accepted
	t.Cleanup(func() { _ = connection.Close() })
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("canceled Open error=%v", err)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Open did not cancel a stalled handshake promptly")
	}

	if err := connection.SetReadDeadline(time.Now().Add(300 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(io.Discard, connection); err != nil {
		var networkError net.Error
		if errors.As(err, &networkError) && networkError.Timeout() {
			t.Error("stalled handshake connection remained open")
		}
	}
}

func TestOpenRejectsPreCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	client, err := Open(ctx, "redis://127.0.0.1:6379", time.Second)
	if client != nil || !errors.Is(err, context.Canceled) {
		t.Errorf("client=%v error=%v", client, err)
	}
}

func TestOpenBoundsStalledHandshakeWithoutCallerDeadline(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	accepted := make(chan net.Conn, 1)
	go func() {
		connection, err := listener.Accept()
		if err == nil {
			accepted <- connection
		}
	}()

	const timeout = 50 * time.Millisecond
	started := time.Now()
	client, err := Open(context.Background(), "redis://"+listener.Addr().String(), timeout)
	elapsed := time.Since(started)
	connection := <-accepted
	_ = connection.Close()
	if client == nil || err == nil {
		t.Fatalf("stalled handshake client=%v error=%v", client, err)
	}
	client.Close()
	if elapsed < timeout/2 || elapsed > 300*time.Millisecond {
		t.Errorf("stalled handshake elapsed=%v, configured timeout=%v", elapsed, timeout)
	}
}

func TestReconnectHandshakeUsesConfiguredBound(t *testing.T) {
	reserved, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := reserved.Addr().String()
	_ = reserved.Close()

	const timeout = 50 * time.Millisecond
	client, err := Open(t.Context(), "redis://"+address, timeout)
	if client == nil || err == nil {
		t.Fatalf("initial client=%v error=%v", client, err)
	}
	t.Cleanup(client.Close)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	accepted := make(chan net.Conn, 1)
	go func() {
		connection, err := listener.Accept()
		if err == nil {
			accepted <- connection
		}
	}()

	started := time.Now()
	result := make(chan error, 1)
	go func() {
		result <- client.Do(context.Background(), client.B().Get().Key("stalled").Build()).Error()
	}()
	connection := <-accepted
	t.Cleanup(func() { _ = connection.Close() })
	err = <-result
	elapsed := time.Since(started)
	if err == nil {
		t.Fatal("reconnect handshake unexpectedly succeeded")
	}
	if elapsed < timeout/2 || elapsed > 300*time.Millisecond {
		t.Errorf("reconnect handshake elapsed=%v, configured timeout=%v", elapsed, timeout)
	}
}

func TestCommandReturnsPromptlyWhenContextIsCanceled(t *testing.T) {
	address, commandReceived, closeServer := startStalledCommandServer(t)
	client, err := Open(t.Context(), "redis://"+address, time.Second)
	if err != nil {
		closeServer()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		closeServer()
		client.Close()
	})

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		result <- client.Do(ctx, client.B().Get().Key("stalled").Build()).Error()
	}()
	<-commandReceived
	cancel()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("canceled command error=%v", err)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("command did not return promptly after context cancellation")
	}
}

func TestEstablishedPipelineUsesConfiguredLivenessBound(t *testing.T) {
	address, commandReceived, closeServer := startStalledCommandServer(t)
	const timeout = 50 * time.Millisecond
	client, err := Open(t.Context(), "redis://"+address, timeout)
	if err != nil {
		closeServer()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		closeServer()
		client.Close()
	})

	started := time.Now()
	result := make(chan error, 1)
	go func() {
		result <- client.Do(context.Background(), client.B().Get().Key("stalled").Build()).Error()
	}()
	<-commandReceived
	select {
	case err := <-result:
		if err == nil {
			t.Error("stalled command unexpectedly succeeded")
		}
		if elapsed := time.Since(started); elapsed < timeout || elapsed > 500*time.Millisecond {
			t.Errorf("stalled command elapsed=%v, configured timeout=%v", elapsed, timeout)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stalled established command exceeded the configured liveness bound")
	}
}

func TestClientOptionPreservesURLConnectionSettings(t *testing.T) {
	option, err := clientOption("rediss://user:secret@valkey.test:6380/3?client_name=tadoku-api", 250*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if len(option.InitAddress) != 1 || option.InitAddress[0] != "valkey.test:6380" {
		t.Errorf("addresses=%v", option.InitAddress)
	}
	if option.Username != "user" || option.Password != "secret" || option.SelectDB != 3 {
		t.Errorf("URL settings user=%q password=%q db=%d", option.Username, option.Password, option.SelectDB)
	}
	if option.TLSConfig == nil || option.TLSConfig.ServerName != "valkey.test" {
		t.Errorf("TLS config=%v", option.TLSConfig)
	}
	if option.ClientName != "tadoku-api" || option.Dialer.Timeout != 250*time.Millisecond || option.ConnWriteTimeout != 250*time.Millisecond {
		t.Errorf("client option=%+v", option)
	}
	if !option.ForceSingleClient || !option.DisableRetry || !option.AlwaysPipelining {
		t.Errorf("required standalone options not set")
	}
}

func TestClientOptionRejectsInvalidConfigurationWithoutLeakingCredentials(t *testing.T) {
	for _, test := range []struct {
		name    string
		rawURL  string
		timeout time.Duration
	}{
		{name: "timeout", rawURL: "redis://127.0.0.1:6379"},
		{name: "multiple addresses", rawURL: "redis://127.0.0.1:6379?addr=127.0.0.1:6380", timeout: time.Second},
		{name: "sentinel", rawURL: "redis://127.0.0.1:6379?master_set=main", timeout: time.Second},
		{name: "empty host", rawURL: "redis://:6379", timeout: time.Second},
		{name: "zero port", rawURL: "redis://127.0.0.1:0", timeout: time.Second},
		{name: "large port", rawURL: "redis://127.0.0.1:99999", timeout: time.Second},
		{name: "negative database", rawURL: "redis://127.0.0.1:6379/-1", timeout: time.Second},
		{name: "empty Unix path", rawURL: "unix://", timeout: time.Second},
		{name: "fragment", rawURL: "redis://127.0.0.1:6379#ignored", timeout: time.Second},
		{name: "unknown option", rawURL: "redis://127.0.0.1:6379?unknown=ignored", timeout: time.Second},
		{name: "malformed option", rawURL: "redis://127.0.0.1:6379?client_name=bad%zz", timeout: time.Second},
		{name: "malformed secret", rawURL: "redis://user:do-not-log-%zz@127.0.0.1:6379", timeout: time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := clientOption(test.rawURL, test.timeout)
			if err == nil {
				t.Fatal("expected an error")
			}
			if strings.Contains(err.Error(), "do-not-log") {
				t.Errorf("error exposed credentials: %v", err)
			}
		})
	}
}

func TestClientOptionAcceptsUnixDatabaseQuery(t *testing.T) {
	option, err := clientOption("unix:///tmp/valkey.sock?db=3", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if option.InitAddress[0] != "/tmp/valkey.sock" || option.SelectDB != 3 {
		t.Errorf("address=%q db=%d", option.InitAddress[0], option.SelectDB)
	}
}

func randomID(t *testing.T) string {
	t.Helper()
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func startTCPProxy(t *testing.T, address, target string) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			downstream, err := listener.Accept()
			if err != nil {
				return
			}
			upstream, err := net.Dial("tcp", target)
			if err != nil {
				_ = downstream.Close()
				continue
			}
			go func() {
				defer downstream.Close()
				defer upstream.Close()
				go func() { _, _ = io.Copy(upstream, downstream); _ = upstream.(*net.TCPConn).CloseWrite() }()
				_, _ = io.Copy(downstream, upstream)
			}()
		}
	}()
	t.Cleanup(func() {
		_ = listener.Close()
	})
	return listener
}

func startStalledCommandServer(t *testing.T) (string, <-chan struct{}, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	commandReceived := make(chan struct{}, 1)
	stop := make(chan struct{})
	var once sync.Once
	closeServer := func() {
		once.Do(func() {
			close(stop)
			_ = listener.Close()
		})
	}

	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer connection.Close()
				reader := bufio.NewReader(connection)
				for {
					command, err := readRESPCommand(reader)
					if err != nil {
						return
					}
					switch strings.ToUpper(command[0]) {
					case "HELLO":
						_, err = io.WriteString(connection, "%2\r\n$5\r\nproto\r\n:3\r\n$7\r\nversion\r\n$5\r\n9.0.0\r\n")
					case "GET":
						select {
						case commandReceived <- struct{}{}:
						default:
						}
						<-stop
						return
					default:
						_, err = io.WriteString(connection, "+OK\r\n")
					}
					if err != nil {
						return
					}
				}
			}()
		}
	}()
	return listener.Addr().String(), commandReceived, closeServer
}

func readRESPCommand(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	count, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")[1:])
	if err != nil {
		return nil, err
	}
	command := make([]string, count)
	for i := range command {
		line, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		length, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")[1:])
		if err != nil {
			return nil, err
		}
		value := make([]byte, length+2)
		if _, err := io.ReadFull(reader, value); err != nil {
			return nil, err
		}
		command[i] = string(value[:length])
	}
	return command, nil
}
