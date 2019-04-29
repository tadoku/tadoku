package domain

import (
	"github.com/srvc/fail"
)

// LanguageCode according to ISO 639-3
type LanguageCode string

// These are all the possible values for LanguageCode
// This list is not yet complete, it's just a start
// Language codes from:
// https://en.wikipedia.org/wiki/Wikipedia:WikiProject_Languages/List_of_ISO_639-3_language_codes_(2019)
const (
	Global     LanguageCode = "GLO" // This is uppercase so it wouldn't collide with Galambu
	Chinese    LanguageCode = "zho"
	Dutch      LanguageCode = "nld"
	English    LanguageCode = "eng"
	French     LanguageCode = "fra"
	German     LanguageCode = "deu"
	Greek      LanguageCode = "ell"
	Irish      LanguageCode = "gle"
	Italian    LanguageCode = "ita"
	Japanese   LanguageCode = "jpn"
	Korean     LanguageCode = "kor"
	Portuguese LanguageCode = "por"
	Russian    LanguageCode = "rus"
	Spanish    LanguageCode = "spa"
	Swedish    LanguageCode = "swe"
	Turkish    LanguageCode = "tur"
)

// LanguageCodes is a collection of language codes
type LanguageCodes []LanguageCode

// AllLanguageCodes is an array with all possible language codes
var AllLanguageCodes = LanguageCodes{
	Global,
	Chinese,
	Dutch,
	English,
	French,
	German,
	Greek,
	Irish,
	Italian,
	Japanese,
	Korean,
	Portuguese,
	Russian,
	Spanish,
	Swedish,
	Turkish,
}

// ContainsLanguage is a helper to figure out if a collection of languages contains the target language
func (codes LanguageCodes) ContainsLanguage(target LanguageCode) bool {
	for _, code := range codes {
		if code == target {
			return true
		}
	}

	return false
}

// ErrInvalidLanguage for when a language is not defined in our app
var ErrInvalidLanguage = fail.New("supplied language is not supported")

// Validate a language code
func (code LanguageCode) Validate() (bool, error) {
	for _, possibleCode := range AllLanguageCodes {
		if possibleCode == code {
			return true, nil
		}
	}

	return false, ErrInvalidLanguage
}
