package identity

import (
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Email struct{ value string }
type DisplayName struct{ value string }
type Password struct{ value string }

type Error struct {
	Code  string
	cause error
}

func (e *Error) Error() string { return e.cause.Error() }
func (e *Error) Unwrap() error { return e.cause }

func ParseEmail(raw string) (Email, error) {
	value := normalizeEmail(raw)
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value || len(value) > 320 {
		return Email{}, &Error{"invalid_email", errx.NewInvalidInputError("email must be a valid address")}
	}
	return Email{value}, nil
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// recipient is the stored address without the spaces and tabs around it, which RFC 5322 reads as folding whitespace
// outside the mailbox. Case and every other character are kept; the email writer decides whether it is a recipient.
func (e Email) recipient() string {
	return strings.Trim(e.value, " \t")
}

func ParseDisplayName(raw string) (DisplayName, error) {
	if !utf8.ValidString(raw) || utf8.RuneCountInString(raw) < 2 || utf8.RuneCountInString(raw) > 32 {
		return DisplayName{}, &Error{
			"invalid_display_name", errx.NewInvalidInputError("display name must contain 2–32 characters"),
		}
	}
	if strings.TrimSpace(raw) == "" {
		return DisplayName{}, &Error{
			"invalid_display_name", errx.NewInvalidInputError("display name must contain a visible character"),
		}
	}
	return DisplayName{raw}, nil
}

func ParsePassword(raw string, email Email) (Password, error) {
	if !utf8.ValidString(raw) || utf8.RuneCountInString(raw) < 8 {
		return Password{}, &Error{
			"invalid_password", errx.NewInvalidInputError("password must contain at least 8 characters"),
		}
	}
	if len(raw) > 72 {
		return Password{}, &Error{
			"invalid_password", errx.NewInvalidInputError("password must not exceed 72 UTF-8 bytes"),
		}
	}
	if email.value == "" || similarPassword(email.value, raw) {
		return Password{}, &Error{
			"invalid_password", errx.NewInvalidInputError("password is too similar to the email"),
		}
	}
	return Password{raw}, nil
}

func similarPassword(email, password string) bool {
	a, b := []rune(strings.ToLower(email)), []rune(strings.ToLower(password))
	distance, common := make([]int, len(b)+1), make([]int, len(b)+1)
	for j := range distance {
		distance[j] = j
	}
	longest := 0
	for i, x := range a {
		diagonal, previousCommon := distance[0], 0
		distance[0] = i + 1
		for j, y := range b {
			above, aboveCommon := distance[j+1], common[j+1]
			cost := 1
			common[j+1] = 0
			if x == y {
				cost = 0
				common[j+1] = previousCommon + 1
				longest = max(longest, common[j+1])
			}
			distance[j+1] = min(distance[j]+1, above+1, diagonal+cost)
			diagonal, previousCommon = above, aboveCommon
		}
	}
	return distance[len(b)] < 5 || longest*2 > len(b)
}

func (e Email) String() string       { return e.value }
func (n DisplayName) String() string { return n.value }
func (Password) String() string      { return "[REDACTED]" }

var hostLabel = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func ValidateOrigin(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || !allowedURL(parsed) || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return &Error{"invalid_origin", errx.NewForbiddenError("origin is not an allowed Tadoku app")}
	}
	return nil
}

func ValidateReturnTo(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || len(raw) > 2048 || !allowedURL(parsed) {
		return &Error{"invalid_return_to", errx.NewInvalidInputError("return_to must use an allowed Tadoku app")}
	}
	return nil
}

func allowedURL(value *url.URL) bool {
	if value == nil || value.Scheme != "https" || value.User != nil || value.Port() != "" || value.Opaque != "" {
		return false
	}
	host := strings.ToLower(value.Host)
	for _, parent := range []string{"tadoku.app", "tadoku.dev.lab"} {
		if host == parent || host == "account."+parent || host == "admin."+parent || host == "authv2."+parent ||
			host == "next."+parent {
			return true
		}
	}
	for _, suffix := range []string{".tadoku.dev.lab", ".preview.tadoku.app"} {
		if !strings.HasSuffix(host, suffix) {
			continue
		}
		labels := strings.Split(strings.TrimSuffix(host, suffix), ".")
		if !hostLabel.MatchString(labels[0]) {
			return false
		}
		if len(labels) == 1 {
			return true
		}
		if len(labels) == 2 && (labels[1] == "account" || labels[1] == "admin" || labels[1] == "authv2" ||
			labels[1] == "next") {
			return true
		}
	}
	return false
}
