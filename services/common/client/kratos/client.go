package kratos

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	kratosapi "github.com/ory/kratos-client-go"
)

var ErrNotFound = errors.New("kratos identity not found")

type Client struct {
	client *kratosapi.APIClient
}

func NewClient(kratosURL string) *Client {
	cfg := kratosapi.NewConfiguration()
	cfg.Servers = kratosapi.ServerConfigurations{{URL: kratosURL}}
	return &Client{client: kratosapi.NewAPIClient(cfg)}
}

func (c *Client) FetchIdentity(ctx context.Context, id uuid.UUID) (*kratosapi.Identity, error) {
	req := c.client.IdentityApi.GetIdentity(ctx, id.String())
	identity, res, err := c.client.IdentityApi.GetIdentityExecute(req)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch identity: %w", err)
	}
	return identity, nil
}

func (c *Client) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := c.FetchIdentity(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) ListIdentities(ctx context.Context, perPage int64, page int64) ([]kratosapi.Identity, error) {
	req := c.client.IdentityApi.ListIdentities(ctx)
	req = req.PerPage(perPage)
	req = req.Page(page)

	identities, _, err := c.client.IdentityApi.ListIdentitiesExecute(req)
	if err != nil {
		return nil, fmt.Errorf("could not list identities: %w", err)
	}
	return identities, nil
}
