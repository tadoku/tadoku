package smtp

import (
	"context"
	"crypto/tls"
	"github.com/google/uuid"
	"time"
)

type Config struct {
	Address, Username, Password, Sender string
	SourceARN, FromARN, ReturnPathARN   string
	Timeout                             time.Duration
	TLSConfig                           *tls.Config
}
type Sender struct{}

func New(Config) (*Sender, error) { return &Sender{}, nil }
func (*Sender) Send(context.Context, uuid.UUID, string, string, string) (string, error) {
	return "", nil
}
