package mail

import (
	netmail "net/mail"
	"strings"
)

type Template string

const (
	AccountRecovery     Template = "account.recovery"
	AccountEmailChanged Template = "account.email_changed"
)

type Message struct {
	Template  Template
	Recipient string
	Params    map[string]string
}

// The email writer renders only addresses that pass; check a stored address with it before relying on mail to it.
func ValidAddress(value string) bool {
	parsed, err := netmail.ParseAddress(value)
	return err == nil && parsed.Address == value && len(value) <= 320 && !strings.ContainsAny(value, "\r\n")
}
