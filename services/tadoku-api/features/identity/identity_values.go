package identity

type Email struct{ value string }
type DisplayName struct{ value string }
type Password struct{ value string }

func ParseEmail(string) (Email, error)              { return Email{}, ErrNotImplemented }
func ParseDisplayName(string) (DisplayName, error)  { return DisplayName{}, ErrNotImplemented }
func ParsePassword(string, Email) (Password, error) { return Password{}, ErrNotImplemented }

func (e Email) String() string       { return e.value }
func (n DisplayName) String() string { return n.value }
func (Password) String() string      { return "[REDACTED]" }
