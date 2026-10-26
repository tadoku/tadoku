package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
)

type Deletion struct{ repository *Repository }

func NewDeletion(repository *Repository) *Deletion { return &Deletion{repository: repository} }

// LockAccount requires the caller's transaction and must come before the users row lock, the order the settings writes
// use. A deletion request then waits for a settings write that already holds the account row, and a settings write
// that waits for the request rechecks its session and finds the account locked.
func (d *Deletion) LockAccount(ctx context.Context, userID uuid.UUID) error {
	return d.repository.lockAccountForDeletion(ctx, userID)
}

// DeleteAccount requires the caller's transaction and a user locked for deletion. It deletes the user's account,
// sessions and recovery tokens and returns the account.deleted notice for the stored address. An address that cannot
// receive mail gets no notice, so it never blocks the deletion.
func (d *Deletion) DeleteAccount(ctx context.Context, userID uuid.UUID, deletedAt time.Time) ([]mail.Message, error) {
	emails, err := d.repository.deleteAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	var notices []mail.Message
	for _, email := range emails {
		if !mail.ValidAddress(email.recipient()) {
			continue
		}
		notices = append(notices, mail.Message{
			Template:  mail.AccountDeleted,
			Recipient: email.recipient(),
			Params:    map[string]string{"deleted_at": deletedAt.Format(time.RFC3339)},
		})
	}
	return notices, nil
}
