package languages_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
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
			if got := errors.Is(err, languages.ErrInvalidLanguage); got != test.wantError {
				t.Errorf("invalid error=%v, want %t", err, test.wantError)
			}
		})
	}
}
