package smtp_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	smtpinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/smtp"
)

func TestSendAuthRejectionPreservesCause(t *testing.T) {
	t.Parallel()

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
	t.Cleanup(func() { _ = listener.Close() })

	go serveAuthRejectSMTP(listener, cert)

	const username = "smtp-user-secret"
	const password = "smtp-pass-secret"
	sender, err := smtpinfra.New(smtpinfra.Config{
		Address:       listener.Addr().String(),
		Username:      username,
		Password:      password,
		Sender:        "no-reply@tadoku.app",
		SourceARN:     "arn:source",
		FromARN:       "arn:from",
		ReturnPathARN: "arn:return",
		Timeout:       time.Second,
		TLSConfig: &tls.Config{
			RootCAs:    roots,
			ServerName: "example.com",
			MinVersion: tls.VersionTLS12,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = sender.Send(
		context.Background(),
		uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		"reader@example.com",
		"subject",
		"body",
	)
	if err == nil {
		t.Fatal("expected authentication failure")
	}
	if !strings.Contains(err.Error(), "SMTP authentication failed") {
		t.Fatalf("missing wrapper text: %v", err)
	}

	var proto *textproto.Error
	if !errors.As(err, &proto) {
		t.Fatalf("expected *textproto.Error cause, got %T: %v", err, err)
	}
	if proto.Code != 535 {
		t.Fatalf("reply code=%d, want 535", proto.Code)
	}
	if strings.Contains(err.Error(), username) || strings.Contains(err.Error(), password) {
		t.Fatalf("error leaked credentials: %v", err)
	}
}

func serveAuthRejectSMTP(listener net.Listener, cert tls.Certificate) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func(conn net.Conn) {
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
					secured := tls.Server(conn, &tls.Config{
						Certificates: []tls.Certificate{cert},
						MinVersion:   tls.VersionTLS12,
					})
					if err := secured.Handshake(); err != nil {
						return
					}
					conn = secured
					peer = textproto.NewConn(conn)
				case strings.HasPrefix(line, "AUTH PLAIN"):
					_ = peer.PrintfLine("535 credentials invalid")
					return
				case line == "QUIT":
					_ = peer.PrintfLine("221 bye")
					return
				default:
					_ = peer.PrintfLine("500 unexpected command")
				}
			}
		}(conn)
	}
}
