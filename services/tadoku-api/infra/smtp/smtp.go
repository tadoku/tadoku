package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime/quotedprintable"
	"net"
	"net/mail"
	stdsmtp "net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Address, Username, Password, Sender string
	SourceARN, FromARN, ReturnPathARN   string
	Timeout                             time.Duration
	TLSConfig                           *tls.Config
}

type Sender struct {
	config Config
	host   string
}

func New(config Config) (*Sender, error) {
	host, port, err := net.SplitHostPort(config.Address)
	if err != nil || host == "" || port == "" {
		return nil, errors.New("SMTP address must contain host and port")
	}
	if config.Timeout <= 0 {
		return nil, errors.New("SMTP timeout must be positive")
	}

	address, err := mail.ParseAddress(config.Sender)
	if err != nil || address.Address != config.Sender || strings.ContainsAny(config.Sender, "\r\n") {
		return nil, errors.New("SMTP sender must be an email address")
	}
	if config.Username == "" || config.Password == "" {
		return nil, errors.New("SMTP username and password are required")
	}
	for _, value := range []string{config.SourceARN, config.FromARN, config.ReturnPathARN} {
		if value == "" || strings.ContainsAny(value, "\r\n") {
			return nil, errors.New("SMTP SES headers are required and cannot contain line breaks")
		}
	}

	if config.TLSConfig == nil {
		config.TLSConfig = &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
	} else {
		config.TLSConfig = config.TLSConfig.Clone()
	}
	return &Sender{config: config, host: host}, nil
}

func (s *Sender) Send(ctx context.Context, id uuid.UUID, recipient, subject, body string) (string, error) {
	parsed, err := mail.ParseAddress(recipient)
	if err != nil || parsed.Address != recipient || strings.ContainsAny(recipient+subject, "\r\n") {
		return "", errors.New("invalid SMTP message headers")
	}

	bounded, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(bounded, "tcp", s.config.Address)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	deadline, _ := bounded.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return "", err
	}
	stop := context.AfterFunc(bounded, func() { _ = conn.Close() })
	defer stop()

	client, err := stdsmtp.NewClient(conn, s.host)
	if err != nil {
		return "", err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); !ok {
		return "", errors.New("SMTP server requires STARTTLS support")
	}
	if err := client.StartTLS(s.config.TLSConfig); err != nil {
		return "", err
	}
	if err := client.Auth(stdsmtp.PlainAuth("", s.config.Username, s.config.Password, s.host)); err != nil {
		return "", errors.New("SMTP authentication failed")
	}
	if err := client.Mail(s.config.Sender); err != nil {
		return "", err
	}
	if err := client.Rcpt(recipient); err != nil {
		return "", err
	}

	command, err := client.Text.Cmd("DATA")
	if err != nil {
		return "", err
	}
	client.Text.StartResponse(command)
	_, _, err = client.Text.ReadResponse(354)
	client.Text.EndResponse(command)
	if err != nil {
		return "", err
	}

	var message bytes.Buffer
	for _, header := range [][2]string{
		{"From", s.config.Sender},
		{"To", recipient},
		{"Subject", subject},
		{"Message-ID", fmt.Sprintf("<%s@tadoku.app>", id)},
		{"X-SES-SOURCE-ARN", s.config.SourceARN},
		{"X-SES-FROM-ARN", s.config.FromARN},
		{"X-SES-RETURN-PATH-ARN", s.config.ReturnPathARN},
		{"MIME-Version", "1.0"},
		{"Content-Type", "text/plain; charset=utf-8"},
		{"Content-Transfer-Encoding", "quoted-printable"},
	} {
		fmt.Fprintf(&message, "%s: %s\r\n", header[0], header[1])
	}

	message.WriteString("\r\n")
	encoder := quotedprintable.NewWriter(&message)
	if _, err := encoder.Write([]byte(body)); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}

	writer := client.Text.DotWriter()
	if _, err := writer.Write(message.Bytes()); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	_, response, err := client.Text.ReadResponse(250)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(response)
	if len(fields) == 0 {
		return "", errors.New("SMTP server returned no provider message ID")
	}
	return fields[len(fields)-1], nil
}
