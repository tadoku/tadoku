package ory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	commonkratos "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type KratosClient struct {
	client *commonkratos.Client
}

func NewKratosClient(kratosURL string) *KratosClient {
	return &KratosClient{
		client: commonkratos.NewClient(kratosURL),
	}
}

type Traits struct {
	DisplayName string `json:"display_name"`
	Email       string
}

func (k *KratosClient) FetchIdentity(ctx context.Context, id uuid.UUID) (*domain.UserTraits, error) {
	identity, err := k.client.FetchIdentity(ctx, id)
	if err != nil {
		if errors.Is(err, commonkratos.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if identity.GetSchemaId() != "user" {
		return nil, fmt.Errorf("unexpected schema %s", identity.GetSchemaId())
	}

	traitsJSON, err := json.Marshal(identity.GetTraits())
	if err != nil {
		return nil, fmt.Errorf("could not fetch identity: %w", err)
	}

	traits := Traits{}
	if err := json.Unmarshal(traitsJSON, &traits); err != nil {
		return nil, fmt.Errorf("could not fetch identity: %w", err)
	}

	return &domain.UserTraits{
		UserDisplayName: traits.DisplayName,
		Email:           traits.Email,
		CreatedAt:       identity.GetCreatedAt(),
	}, nil
}

func (k *KratosClient) ListIdentities(ctx context.Context, perPage int64, page int64) (*domain.ListIdentitiesResult, error) {
	identities, err := k.client.ListIdentities(ctx, perPage, page)
	if err != nil {
		return nil, err
	}

	result := &domain.ListIdentitiesResult{
		Identities: make([]domain.IdentityInfo, 0, len(identities)),
		HasMore:    len(identities) == int(perPage),
	}

	for _, identity := range identities {
		if identity.GetSchemaId() != "user" {
			continue
		}

		traitsJSON, err := json.Marshal(identity.GetTraits())
		if err != nil {
			continue
		}

		traits := Traits{}
		if err := json.Unmarshal(traitsJSON, &traits); err != nil {
			continue
		}

		createdAt := ""
		if identity.CreatedAt != nil {
			createdAt = identity.GetCreatedAt().Format("2006-01-02T15:04:05Z")
		}

		result.Identities = append(result.Identities, domain.IdentityInfo{
			ID:          identity.GetId(),
			DisplayName: traits.DisplayName,
			Email:       traits.Email,
			CreatedAt:   createdAt,
		})
	}

	return result, nil
}
