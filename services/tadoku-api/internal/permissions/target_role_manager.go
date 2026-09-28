package permissions

import (
	"context"

	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
)

type KetoManager struct {
	keto      ketoclient.AuthorizationClient
	namespace string
	object    string
}

func NewKetoManager(keto ketoclient.AuthorizationClient, namespace, object string) *KetoManager {
	return &KetoManager{
		keto:      keto,
		namespace: namespace,
		object:    object,
	}
}

func (m *KetoManager) SetBanned(ctx context.Context, subjectID string, enabled bool) error {
	if enabled {
		return m.keto.AddRelation(ctx, m.namespace, m.object, "banned", ketoclient.Subject{ID: subjectID})
	}
	return m.keto.DeleteRelation(ctx, m.namespace, m.object, "banned", ketoclient.Subject{ID: subjectID})
}
