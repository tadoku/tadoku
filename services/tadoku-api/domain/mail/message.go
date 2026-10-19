package mail

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
