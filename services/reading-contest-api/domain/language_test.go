package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguage_AllLanguagesDoesNotContainGlobal(t *testing.T) {
	for _, l := range AllLanguages {
		assert.NotEqual(t, Global, l.Code, "language database should not contain a language with the global code")
	}
}
