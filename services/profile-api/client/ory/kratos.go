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
	return NewKratosClientFromClient(commonkratos.NewClient(kratosURL))
}

func NewKratosClientFromClient(client *commonkratos.Client) *KratosClient {
	return &KratosClient{client: client}
}

type Traits struct {
	DisplayName string `json:"display_name"`
	Email       string
}

func (k *KratosClient) ListIdentities(ctx context.Context, pageSize int64, pageToken string) ([]domain.IdentityInfo, string, error) {
	identities, nextPageToken, err := k.client.ListIdentities(ctx, pageSize, pageToken)
	if err != nil {
		return nil, "", err
	}

	result := make([]domain.IdentityInfo, 0, len(identities))

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

		result = append(result, domain.IdentityInfo{
			ID:          identity.GetId(),
			DisplayName: traits.DisplayName,
			Email:       traits.Email,
			CreatedAt:   createdAt,
		})
	}

	return result, nextPageToken, nil
}

// Verify KratosClient implements domain.KratosClient at compile time.
var _ domain.KratosClient = (*KratosClient)(nil)
