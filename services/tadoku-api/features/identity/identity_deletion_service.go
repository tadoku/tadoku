package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
)

type Deletion struct{ repository *Repository }

func NewDeletion(repository *Repository) *Deletion { return &Deletion{repository: repository} }

func (d *Deletion) DeleteAccount(ctx context.Context, userID uuid.UUID, deletedAt time.Time) ([]mail.Message, error) {
	return nil, nil
}
