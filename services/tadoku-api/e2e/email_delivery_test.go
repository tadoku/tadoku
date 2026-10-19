package e2e_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	smtpinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/smtp"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type emailSMTP struct {
	listener net.Listener
	config   smtpinfra.Config
	mu       sync.Mutex
	messages []string
	attempts int
	reject   bool
}

func emailSMTPServer(t *testing.T, reject bool) *emailSMTP {
	t.Helper()
	certServer := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	cert := certServer.TLS.Certificates[0]
	parsed, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(parsed)
	certServer.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	fixture := &emailSMTP{listener: listener, reject: reject, config: smtpinfra.Config{
		Address: listener.Addr().String(), Username: "fixture", Password: "fixture", Sender: "no-reply@tadoku.app",
		SourceARN: "arn:source", FromARN: "arn:from", ReturnPathARN: "arn:return", Timeout: time.Second,
		TLSConfig: &tls.Config{RootCAs: roots, ServerName: "example.com", MinVersion: tls.VersionTLS12},
	}}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go fixture.serve(conn, cert)
		}
	}()
	return fixture
}

func (s *emailSMTP) serve(conn net.Conn, cert tls.Certificate) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	peer := textproto.NewConn(conn)
	_ = peer.PrintfLine("220 fixture ESMTP")
	for {
		line, err := peer.ReadLine()
		if err != nil {
			return
		}
		switch {
		case strings.HasPrefix(line, "EHLO"):
			_, _ = io.WriteString(conn, "250-fixture\r\n250-STARTTLS\r\n250 AUTH PLAIN\r\n")
		case line == "STARTTLS":
			_ = peer.PrintfLine("220 go ahead")
			secured := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
			if err := secured.Handshake(); err != nil {
				return
			}
			conn = secured
			peer = textproto.NewConn(conn)
		case strings.HasPrefix(line, "AUTH PLAIN"):
			_ = peer.PrintfLine("235 authenticated")
		case strings.HasPrefix(line, "MAIL FROM"), strings.HasPrefix(line, "RCPT TO"):
			_ = peer.PrintfLine("250 ok")
		case line == "DATA":
			_ = peer.PrintfLine("354 data")
			raw, err := peer.ReadDotBytes()
			if err != nil {
				return
			}
			s.mu.Lock()
			s.attempts++
			if !s.reject {
				s.messages = append(s.messages, string(raw))
			}
			s.mu.Unlock()
			if s.reject {
				_ = peer.PrintfLine("451 temporarily unavailable")
			} else {
				_ = peer.PrintfLine("250 Ok fixture-message-id")
			}
		case line == "QUIT":
			_ = peer.PrintfLine("221 bye")
			return
		default:
			_ = peer.PrintfLine("500 unexpected command")
		}
	}
}

func TestEmailWorkerDelivery(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		enabled, production, reject bool
		status                      string
		attempts                    int
	}{
		{"disabled production", false, true, false, "held", 0},
		{"disabled test", false, false, false, "held", 0},
		{"enabled test", true, false, false, "held", 0},
		{"enabled production", true, true, false, "sent", 1},
		{"final SMTP failure", true, true, true, "failed", 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api.reset(t, "")
			fixture := emailSMTPServer(t, tc.reject)
			var sender email.Sender
			if tc.enabled {
				var err error
				sender, err = smtpinfra.New(fixture.config)
				if err != nil {
					t.Fatal(err)
				}
			}
			service := email.NewPersistentService(email.NewRepository(api.db.AppPool), sender)
			key := tenant.Production()
			if !tc.production {
				key, _ = tenant.Parse("e2e/email-worker")
				if _, err := api.db.Pool.Exec(t.Context(), "insert into tenants(key,kind) values ($1,'test')", key.String()); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_, _ = api.db.Pool.Exec(context.Background(), "delete from tenants where key=$1", key.String())
				})
			}
			ctx := tenant.WithKey(t.Context(), key)
			id := writeEmail(t, ctx, service, recoveryMail())
			scope, _ := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
			if !tc.production {
				scope, _ = jobqueue.OnlyTenant(key)
			}
			application, err := worker.NewApplication(jobqueue.NewService(jobqueue.NewRepository(api.db.AppPool)), api.leaderboard, worker.Config{
				Email: service, Scope: scope, Concurrency: 1, ShutdownTimeout: time.Second, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			})
			if err != nil {
				t.Fatal(err)
			}
			workerCtx, cancel := context.WithCancel(t.Context())
			done := make(chan error, 1)
			go func() { done <- application.Run(workerCtx) }()
			defer func() {
				cancel()
				if err := <-done; err != nil {
					t.Error(err)
				}
			}()
			deadline := time.Now().Add(12 * time.Second)
			for {
				var status, state string
				var attempts int
				err := api.db.Pool.QueryRow(ctx, `select e.status,j.state,j.attempts from emails e join jobs j on j.tenant=e.tenant and j.payload->>'email_id'=e.id::text where e.id=$1`, id).Scan(&status, &state, &attempts)
				if err != nil {
					t.Fatal(err)
				}
				if status == tc.status && (state == "completed" || state == "failed") {
					break
				}
				if tc.reject && attempts > 0 && attempts < 5 && status != "pending" {
					t.Fatalf("email terminal before final attempt: status=%s attempts=%d", status, attempts)
				}
				if time.Now().After(deadline) {
					t.Fatalf("delivery timeout status=%s job=%s attempts=%d", status, state, attempts)
				}
				if tc.reject {
					_, err = api.db.Pool.Exec(ctx, "update jobs set next_attempt_at=now()-interval '1 minute' where state='pending'")
					if err != nil {
						t.Fatal(err)
					}
				}
				time.Sleep(20 * time.Millisecond)
			}
			cancel()
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			done <- nil
			row, err := service.Find(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if tc.status == "sent" {
				sent, ok := row.Delivery.(email.Sent)
				if !ok || sent.At.IsZero() || sent.ProviderMessageID != "fixture-message-id" {
					t.Fatalf("delivery=%+v", row.Delivery)
				}
			}
			if tc.status == "failed" {
				failed, ok := row.Delivery.(email.Failed)
				if !ok || failed.Error == "" {
					t.Fatalf("delivery=%+v", row.Delivery)
				}
			}
			if err := service.Send(ctx, id); err != nil {
				t.Fatal(err)
			}
			fixture.mu.Lock()
			attempts := fixture.attempts
			messages := append([]string(nil), fixture.messages...)
			fixture.mu.Unlock()
			if attempts != tc.attempts {
				t.Errorf("SMTP attempts=%d want %d", attempts, tc.attempts)
			}
			if tc.status == "sent" {
				message, err := textproto.NewReader(bufio.NewReader(strings.NewReader(messages[0]))).ReadMIMEHeader()
				if err != nil {
					t.Fatal(err)
				}
				for header, want := range map[string]string{"From": "no-reply@tadoku.app", "To": "reader@example.test", "X-SES-SOURCE-ARN": "arn:source", "X-SES-FROM-ARN": "arn:from", "X-SES-RETURN-PATH-ARN": "arn:return", "Message-ID": fmt.Sprintf("<%s@tadoku.app>", id)} {
					if message.Get(header) != want {
						t.Errorf("header %s=%q want %q", header, message.Get(header), want)
					}
				}
			}
		})
	}
}

func TestEmailGuessedJobTenant(t *testing.T) {
	api.reset(t, "")
	fixture := emailSMTPServer(t, false)
	sender, err := smtpinfra.New(fixture.config)
	if err != nil {
		t.Fatal(err)
	}
	service := email.NewPersistentService(email.NewRepository(api.db.AppPool), sender)
	productionCtx := tenant.WithKey(t.Context(), tenant.Production())
	id := writeEmail(t, productionCtx, service, recoveryMail())
	key, _ := tenant.Parse("e2e/email-guess")
	if _, err := api.db.Pool.Exec(t.Context(), "insert into tenants(key,kind) values ($1,'test')", key.String()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = api.db.Pool.Exec(context.Background(), "delete from tenants where key=$1", key.String())
	}()
	if err := service.Send(tenant.WithKey(t.Context(), key), id); err != nil {
		t.Fatal(err)
	}
	row, err := service.Find(productionCtx, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := row.Delivery.(email.Pending); !ok {
		t.Errorf("guessed job changed production row: %T", row.Delivery)
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if fixture.attempts != 0 {
		t.Errorf("guessed job made %d SMTP attempts", fixture.attempts)
	}
}
