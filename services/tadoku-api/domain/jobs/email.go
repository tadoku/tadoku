package jobs

import (
	"errors"

	"github.com/google/uuid"
)

type SendEmailV1 struct {
	EmailID uuid.UUID `json:"email_id"`
}

const EmailSendV1 Type = "email.send.v1"

func (SendEmailV1) Type() Type { return EmailSendV1 }
func (SendEmailV1) job()       {}
func (j SendEmailV1) Validate() error {
	if j.EmailID == uuid.Nil {
		return errors.New("email_id must be a nonzero UUID")
	}
	return nil
}
