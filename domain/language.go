package domain

// Language contains a language Name and code
type Language struct {
	Code LanguageCode
	Name string
}

// Languages is a collection of languages
type Languages []Language

// AllLanguages is an array with all possible languages
var AllLanguages = Languages{
	Language{Code: Arabic, Name: "Arabic"},
	Language{Code: Chinese, Name: "Chinese"},
	Language{Code: Czech, Name: "Czech"},
	Language{Code: Dutch, Name: "Dutch"},
	Language{Code: Esperanto, Name: "Esperanto"},
	Language{Code: English, Name: "English"},
	Language{Code: French, Name: "French"},
	Language{Code: German, Name: "German"},
	Language{Code: Greek, Name: "Greek"},
	Language{Code: Irish, Name: "Irish"},
	Language{Code: Italian, Name: "Italian"},
	Language{Code: Japanese, Name: "Japanese"},
	Language{Code: Korean, Name: "Korean"},
	Language{Code: Polish, Name: "Polish"},
	Language{Code: Portuguese, Name: "Portuguese"},
	Language{Code: Russian, Name: "Russian"},
	Language{Code: Spanish, Name: "Spanish"},
	Language{Code: Swedish, Name: "Swedish"},
	Language{Code: Thai, Name: "Thai"},
	Language{Code: Turkish, Name: "Turkish"},
}
