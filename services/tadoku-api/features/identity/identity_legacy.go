package identity

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	kratosapi "github.com/ory/kratos-client-go"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity/kratos"
)

var ErrLegacyNotFound = kratos.ErrNotFound

type LegacyReader interface {
	FetchIdentity(context.Context, uuid.UUID) (*kratosapi.Identity, error)
	UserExists(context.Context, uuid.UUID) (bool, error)
	ListIdentities(context.Context, int64, string) ([]kratosapi.Identity, string, error)
}

func NewLegacyReader(url string, options ...kratos.Option) LegacyReader {
	return kratos.NewClient(url, options...)
}

func WithHTTPClient(client *http.Client) kratos.Option { return kratos.WithHTTPClient(client) }
