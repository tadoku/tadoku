package domain

import (
	"strings"

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
	Czech      LanguageCode = "ces"
	Dutch      LanguageCode = "nld"
	English    LanguageCode = "eng"
	Esperanto  LanguageCode = "epo"
	French     LanguageCode = "fra"
	German     LanguageCode = "deu"
	Greek      LanguageCode = "ell"
	Irish      LanguageCode = "gle"
	Italian    LanguageCode = "ita"
	Japanese   LanguageCode = "jpn"
	Korean     LanguageCode = "kor"
	Polish     LanguageCode = "pol"
	Portuguese LanguageCode = "por"
	Russian    LanguageCode = "rus"
	Spanish    LanguageCode = "spa"
	Swedish    LanguageCode = "swe"
	Thai       LanguageCode = "tha"
	Turkish    LanguageCode = "tur"
)

// LanguageCodes is a collection of language codes
type LanguageCodes []LanguageCode

// AllLanguageCodes is an array with all possible language codes
var AllLanguageCodes = LanguageCodes{
	Global,
	Chinese,
	Czech,
	Dutch,
	Esperanto,
	English,
	French,
	German,
	Greek,
	Irish,
	Italian,
	Japanese,
	Korean,
	Polish,
	Portuguese,
	Russian,
	Spanish,
	Swedish,
	Thai,
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

// Len is the number of elements in the collection.
func (codes LanguageCodes) Len() int {
	return len(codes)
}

// Swap swaps the elements with indexes i and j.
func (codes LanguageCodes) Swap(i, j int) {
	codes[i], codes[j] = codes[j], codes[i]
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (codes LanguageCodes) Less(i, j int) bool {
	return strings.Compare(string(codes[i]), string(codes[j])) >= 0
}
