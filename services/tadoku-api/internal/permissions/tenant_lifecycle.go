package permissions

import (
	"context"

	"github.com/google/uuid"
	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type TenantManager struct {
	keto *ketoclient.Client
}

func NewTenantManager(keto *ketoclient.Client) *TenantManager {
	return &TenantManager{keto: keto}
}

func (manager *TenantManager) Provision(ctx context.Context, key tenant.TestKey, testers []uuid.UUID) error {
	object, err := rolesObject(tenant.WithKey(ctx, key.Key()))
	if err != nil {
		return err
	}
	parent, err := rolesObject(tenant.WithKey(ctx, tenant.Production()))
	if err != nil {
		return err
	}

	subject := ketoclient.Subject{Set: &ketoclient.SubjectSet{Namespace: "app", Object: parent}}
	if err := manager.ensureRelation(ctx, object, "parents", subject); err != nil {
		return err
	}
	for _, tester := range testers {
		err := manager.ensureRelation(ctx, object, "testers", ketoclient.Subject{ID: tester.String()})
		if err != nil {
			return err
		}
	}
	return nil
}

func (manager *TenantManager) ensureRelation(ctx context.Context, object, relation string, subject ketoclient.Subject) error {
	exists, err := manager.keto.HasRelation(ctx, "app", object, relation, subject)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return manager.keto.AddRelation(ctx, "app", object, relation, subject)
}

func (manager *TenantManager) Delete(ctx context.Context, key tenant.TestKey) error {
	object, err := rolesObject(tenant.WithKey(ctx, key.Key()))
	if err != nil {
		return err
	}
	return manager.keto.DeleteObjectRelations(ctx, "app", object)
}
