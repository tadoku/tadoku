package ory

import (
	"context"
	"encoding/json"

	commonkratos "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/profile-api/domain"
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

// Verify KratosClient implements domain.KratosClient at compile time.
var _ domain.KratosClient = (*KratosClient)(nil)
