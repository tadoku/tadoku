package roles

import (
	"context"

	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
)

// TODO: This should not be in common but rather in profile-api when the role management endpoints live there
type Manager interface {
	SetAdmin(ctx context.Context, subjectID string, enabled bool) error
	SetBanned(ctx context.Context, subjectID string, enabled bool) error
}

type KetoManager struct {
	keto      ketoclient.RelationManager
	namespace string
	object    string
}

func NewKetoManager(keto ketoclient.RelationManager, namespace, object string) *KetoManager {
	return &KetoManager{
		keto:      keto,
		namespace: namespace,
		object:    object,
	}
}

func (m *KetoManager) SetAdmin(ctx context.Context, subjectID string, enabled bool) error {
	if enabled {
		return m.keto.AddRelation(ctx, m.namespace, m.object, "admins", ketoclient.Subject{ID: subjectID})
	}
	return m.keto.DeleteRelation(ctx, m.namespace, m.object, "admins", ketoclient.Subject{ID: subjectID})
}

func (m *KetoManager) SetBanned(ctx context.Context, subjectID string, enabled bool) error {
	if enabled {
		return m.keto.AddRelation(ctx, m.namespace, m.object, "banned", ketoclient.Subject{ID: subjectID})
	}
	return m.keto.DeleteRelation(ctx, m.namespace, m.object, "banned", ketoclient.Subject{ID: subjectID})
}
