package jobs

import (
	"errors"

	"github.com/google/uuid"
)

const AccountDeleteV1 Type = "account.delete.v1"

// DeleteAccountV1 deletes the data of a user whose deletion is already requested. It reads current state, so a job for
// a user who is already deleted completes without changes and replay is safe. There is no cancel path.
type DeleteAccountV1 struct {
	UserID uuid.UUID `json:"user_id"`
}

func (DeleteAccountV1) Type() Type { return AccountDeleteV1 }
func (DeleteAccountV1) job()       {}
func (j DeleteAccountV1) Validate() error {
	if j.UserID == uuid.Nil {
		return errors.New("user_id must be a nonzero UUID")
	}
	return nil
}
