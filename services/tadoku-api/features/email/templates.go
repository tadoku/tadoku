package email

import (
	"bytes"
	"net/url"
	"strings"
	"text/template"
	"time"

	domainmail "github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type emailTemplate struct {
	params  []string
	subject string
	body    *template.Template
}

var templates = map[domainmail.Template]emailTemplate{
	domainmail.AccountRecovery: {
		params:  []string{"recovery_url"},
		subject: "Recover your Tadoku account",
		body: template.Must(template.New("recovery").Parse(
			"Recover your Tadoku account using this link:\n\n{{.recovery_url}}\n\n" +
				"This link expires in one hour. If you did not request it, you can ignore this email.\n",
		)),
	},
	domainmail.AccountEmailChanged: {
		params:  []string{"new_email", "changed_at"},
		subject: "Your Tadoku email address changed",
		body: template.Must(template.New("changed").Parse(
			"Your Tadoku email address changed to {{.new_email}} at {{.changed_at}}.\n\n" +
				"If you did not make this change, recover your account or contact Tadoku support.\n",
		)),
	},
}

func render(message domainmail.Message) (string, string, error) {
	entry, ok := templates[message.Template]
	if !ok {
		return "", "", errx.NewInvalidInputError("unknown email template")
	}
	if !domainmail.ValidAddress(message.Recipient) {
		return "", "", errx.NewInvalidInputError("invalid email recipient")
	}
	for _, name := range entry.params {
		if strings.TrimSpace(message.Params[name]) == "" {
			return "", "", errx.NewInvalidInputError("missing email parameter: " + name)
		}
		if len(message.Params[name]) > 4096 {
			return "", "", errx.NewInvalidInputError("email parameter exceeds 4096 bytes: " + name)
		}
	}
	if len(message.Params) != len(entry.params) {
		return "", "", errx.NewInvalidInputError("unexpected email parameters")
	}
	switch message.Template {
	case domainmail.AccountRecovery:
		if !allowedRecoveryURL(message.Params["recovery_url"]) {
			return "", "", errx.NewInvalidInputError("recovery_url must use an allowed Tadoku host")
		}
	case domainmail.AccountEmailChanged:
		if !domainmail.ValidAddress(message.Params["new_email"]) {
			return "", "", errx.NewInvalidInputError("invalid new_email parameter")
		}
		if _, err := time.Parse(time.RFC3339, message.Params["changed_at"]); err != nil {
			return "", "", errx.NewInvalidInputError("changed_at must be an RFC3339 timestamp")
		}
	}
	var body bytes.Buffer
	if err := entry.body.Execute(&body, message.Params); err != nil {
		return "", "", err
	}
	return entry.subject, body.String(), nil
}

func allowedRecoveryURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	for _, base := range []string{"tadoku.app", "tadoku.dev.lab"} {
		if host == base || host == "account."+base || host == "admin."+base || host == "authv2."+base {
			return true
		}
		if strings.HasSuffix(host, ".preview."+base) {
			prefix := strings.TrimSuffix(host, ".preview."+base)
			for _, part := range strings.Split(prefix, ".") {
				if part == "" {
					return false
				}
			}
			return true
		}
		if base == "tadoku.dev.lab" && strings.HasSuffix(host, "."+base) {
			prefix := strings.TrimSuffix(host, "."+base)
			labels := strings.Split(prefix, ".")
			if len(labels) == 1 && labels[0] != "" {
				return true
			}
			if len(labels) == 2 && labels[0] != "" && (labels[1] == "account" || labels[1] == "admin" || labels[1] == "authv2") {
				return true
			}
		}
	}
	return false
}
