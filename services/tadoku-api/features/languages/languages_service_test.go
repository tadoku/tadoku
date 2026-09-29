package languages_test

import (
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestLanguageParametersValidateByteLengths(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		code      string
		langName  string
		wantError bool
	}{
		{name: "minimum", code: "x", langName: "x"},
		{name: "maximum", code: strings.Repeat("x", 10), langName: strings.Repeat("x", 100)},
		{name: "whitespace retained", code: " x ", langName: " x "},
		{name: "empty code", langName: "x", wantError: true},
		{name: "empty name", code: "x", wantError: true},
		{name: "long code", code: strings.Repeat("x", 11), langName: "x", wantError: true},
		{name: "long name", code: "x", langName: strings.Repeat("x", 101), wantError: true},
		{name: "unicode code bytes", code: "日本語語", langName: "x", wantError: true},
		{name: "unicode name bytes", code: "x", langName: strings.Repeat("語", 34), wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			createErr := languages.CreateLanguageParameters{Code: test.code, Name: test.langName}.Validate()
			if got := errx.KindOf(createErr) == errx.InvalidInput; got != test.wantError {
				t.Errorf("CreateLanguageParameters: invalid error=%v, want %t", createErr, test.wantError)
			}
			updateErr := languages.UpdateLanguageParameters{Code: test.code, Name: test.langName}.Validate()
			if got := errx.KindOf(updateErr) == errx.InvalidInput; got != test.wantError {
				t.Errorf("UpdateLanguageParameters: invalid error=%v, want %t", updateErr, test.wantError)
			}
		})
	}
}
