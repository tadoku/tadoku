package ory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	kratos "github.com/ory/kratos-client-go"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

type KratosClient struct {
	client *kratos.APIClient
}

func NewKratosClient(kratosURL string) *KratosClient {
	cfg := kratos.NewConfiguration()
	cfg.Servers = kratos.ServerConfigurations{{URL: kratosURL}}

	return &KratosClient{
		client: kratos.NewAPIClient(cfg),
	}
}

type Traits struct {
	DisplayName string `json:"display_name"`
	Email       string
}

func (k *KratosClient) FetchIdentity(ctx context.Context, id uuid.UUID) (*query.UserTraits, error) {
	req := k.client.IdentityApi.GetIdentity(ctx, id.String())
	identity, res, err := k.client.IdentityApi.GetIdentityExecute(req)
	if err != nil {
		if res.StatusCode == http.StatusNotFound {
			return nil, query.ErrNotFound
		}
		return nil, fmt.Errorf("could not fetch identity: %w", err)
	}

	if identity.SchemaId != "user" {
		return nil, fmt.Errorf("unexpected schema %s", identity.SchemaId)
	}

	traitsJSON, err := json.Marshal(identity.Traits)
	if err != nil {
		return nil, fmt.Errorf("could not fetch identity: %w", err)
	}

	traits := Traits{}
	if err := json.Unmarshal(traitsJSON, &traits); err != nil {
		return nil, fmt.Errorf("could not fetch identity: %w", err)
	}

	return &query.UserTraits{
		UserDisplayName: traits.DisplayName,
		Email:           traits.Email,
		CreatedAt:       *identity.CreatedAt,
	}, nil
}
