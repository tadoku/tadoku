package ory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	kratos "github.com/ory/kratos-client-go"
)

type KratosClient struct {
	client *kratos.APIClient
}

func NewKratosClient(kratosURL string) *KratosClient {
	cfg := kratos.NewConfiguration()
	cfg.Servers = kratos.ServerConfigurations{{URL: kratosURL}}
	return &KratosClient{client: kratos.NewAPIClient(cfg)}
}

func (k *KratosClient) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	req := k.client.IdentityApi.GetIdentity(ctx, id.String())
	_, res, err := k.client.IdentityApi.GetIdentityExecute(req)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, fmt.Errorf("could not fetch identity: %w", err)
	}
	return true, nil
}

var _ interface {
	UserExists(ctx context.Context, id uuid.UUID) (bool, error)
} = (*KratosClient)(nil)
