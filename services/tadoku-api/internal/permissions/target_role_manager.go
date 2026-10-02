package permissions

import (
	"context"

	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
)

type KetoManager struct {
	keto ketoclient.AuthorizationClient
}

func NewKetoManager(keto ketoclient.AuthorizationClient) *KetoManager {
	return &KetoManager{keto: keto}
}

func (m *KetoManager) SetBanned(ctx context.Context, subjectID string, enabled bool) error {
	object, err := rolesObject(ctx)
	if err != nil {
		return err
	}

	if enabled {
		return m.keto.AddRelation(ctx, "app", object, "banned", ketoclient.Subject{ID: subjectID})
	}
	return m.keto.DeleteRelation(ctx, "app", object, "banned", ketoclient.Subject{ID: subjectID})
}
