package languages_test

import (
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestCreateLanguageParametersValidateByteLengths(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		parameters languages.CreateLanguageParameters
		wantError  bool
	}{
		{name: "minimum", parameters: languages.CreateLanguageParameters{Code: "x", Name: "x"}},
		{name: "maximum", parameters: languages.CreateLanguageParameters{Code: strings.Repeat("x", 10), Name: strings.Repeat("x", 100)}},
		{name: "whitespace retained", parameters: languages.CreateLanguageParameters{Code: " x ", Name: " x "}},
		{name: "empty code", parameters: languages.CreateLanguageParameters{Name: "x"}, wantError: true},
		{name: "empty name", parameters: languages.CreateLanguageParameters{Code: "x"}, wantError: true},
		{name: "long code", parameters: languages.CreateLanguageParameters{Code: strings.Repeat("x", 11), Name: "x"}, wantError: true},
		{name: "long name", parameters: languages.CreateLanguageParameters{Code: "x", Name: strings.Repeat("x", 101)}, wantError: true},
		{name: "unicode code bytes", parameters: languages.CreateLanguageParameters{Code: "日本語語", Name: "x"}, wantError: true},
		{name: "unicode name bytes", parameters: languages.CreateLanguageParameters{Code: "x", Name: strings.Repeat("語", 34)}, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.parameters.Validate()
			if got := errx.KindOf(err) == errx.InvalidInput; got != test.wantError {
				t.Errorf("invalid error=%v, want %t", err, test.wantError)
			}
		})
	}
}

func TestUpdateLanguageParametersValidateByteLengths(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		value     string
		wantError bool
	}{
		{name: "minimum", value: "x"},
		{name: "maximum", value: strings.Repeat("x", 100)},
		{name: "whitespace retained", value: " x "},
		{name: "empty", wantError: true},
		{name: "long", value: strings.Repeat("x", 101), wantError: true},
		{name: "unicode bytes", value: strings.Repeat("語", 34), wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := (languages.UpdateLanguageParameters{Code: "code", Name: test.value}).Validate()
			if got := errx.KindOf(err) == errx.InvalidInput; got != test.wantError {
				t.Errorf("invalid error=%v, want %t", err, test.wantError)
			}
		})
	}
}
