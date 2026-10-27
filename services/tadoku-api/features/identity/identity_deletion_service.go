package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
)

type Deletion struct{ repository *Repository }

func NewDeletion(repository *Repository) *Deletion { return &Deletion{repository: repository} }

// LockAccountForRequest requires the caller's transaction and comes before the users row lock: every transaction
// that locks both rows takes the account row first. It skips a user whose deletion was already filed when the
// statement started.
func (d *Deletion) LockAccountForRequest(ctx context.Context, userID uuid.UUID) error {
	return d.repository.lockAccountForRequest(ctx, userID)
}

// LockAccountForDeletion requires the worker's deletion transaction and comes before the users row lock, so the
// worker waits for any login, recovery or settings write that holds the account row instead of deadlocking with it.
func (d *Deletion) LockAccountForDeletion(ctx context.Context, userID uuid.UUID) error {
	return d.repository.lockAccountForDeletion(ctx, userID)
}

// ConfirmAccount requires the caller's transaction to hold the account row from LockAccountForRequest, so the address
// cannot change before the transaction ends. Unless email, normalized as login normalizes it, is the account's
// address, it returns ErrConfirmationMismatch.
func (d *Deletion) ConfirmAccount(ctx context.Context, userID uuid.UUID, email string) error {
	matches, err := d.repository.accountHasEmail(ctx, userID, normalizeEmail(email))
	if err != nil {
		return err
	}
	if !matches {
		return ErrConfirmationMismatch
	}
	return nil
}

// DeleteAccount requires the caller's transaction and a user locked for deletion. It deletes the user's account,
// sessions and recovery tokens and returns the account.deleted notice for the stored address. It returns no notice
// when the user has no account or the address cannot receive mail, so an address never blocks the deletion.
func (d *Deletion) DeleteAccount(
	ctx context.Context,
	userID uuid.UUID,
	deletedAt time.Time,
) (mail.Message, bool, error) {
	email, found, err := d.repository.deleteAccount(ctx, userID)
	if err != nil || !found {
		return mail.Message{}, false, err
	}
	if !mail.ValidAddress(email.recipient()) {
		return mail.Message{}, false, nil
	}

	return mail.Message{
		Template:  mail.AccountDeleted,
		Recipient: email.recipient(),
		Params:    map[string]string{"deleted_at": deletedAt.Format(time.RFC3339)},
	}, true, nil
}
