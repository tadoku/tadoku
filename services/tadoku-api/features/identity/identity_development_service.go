package identity

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"golang.org/x/crypto/bcrypt"
)

type DevelopmentAccount struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Password    string
}

// Call only from development provisioning inside its authorized test-tenant transaction.
// Existing fixture UUIDs belong to the branch; reseeding preserves their account data.
func SeedDevelopmentAccounts(
	ctx context.Context,
	repository *Repository,
	key tenant.TestKey,
	fixtures []DevelopmentAccount,
) error {
	current, ok := tenant.FromContext(ctx)
	if !ok || key.String() == "" || current != key.Key() {
		return errors.New("development accounts require their parsed test tenant")
	}
	for _, fixture := range fixtures {
		email, err := ParseEmail(fixture.Email)
		if err != nil {
			return err
		}
		name, err := ParseDisplayName(fixture.DisplayName)
		if err != nil {
			return err
		}
		if fixture.ID == uuid.Nil || len(fixture.Password) == 0 || len(fixture.Password) > 72 {
			return errors.New("development fixture requires an ID and a password of 1–72 bytes")
		}

		exists, err := repository.developmentAccountExists(ctx, fixture.ID)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(fixture.Password), 12)
		if err != nil {
			return errors.New("hash development fixture password")
		}
		if err := repository.seedDevelopmentUser(ctx, fixture.ID, name); err != nil {
			return err
		}
		if err := repository.seedDevelopmentAccount(ctx, fixture.ID, email, string(hash)); err != nil {
			return err
		}
	}
	return nil
}
