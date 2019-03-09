package domain

// LanguageCode according to ISO 639-3
type LanguageCode string

// These are all the possible values for LanguageCode
// This list is not yet complete, it's just a start
const (
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

// AllLanguages is an array with all possible languages
var AllLanguages = LanguageCodes{
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
