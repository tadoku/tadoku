package domain

const (
	UnitKeyReadingPage           = "reading_page"
	UnitKeyReadingTwoColumnPage  = "reading_two_column_page"
	UnitKeyReadingComicPage      = "reading_comic_page"
	UnitKeyReadingSentence       = "reading_sentence"
	UnitKeyReadingCharacter      = "reading_character"
	UnitKeyListeningMinute       = "listening_minute"
	UnitKeyListeningDenseMinutes = "listening_dense_minutes"
	UnitKeyWritingPage           = "writing_page"
	UnitKeyWritingSentence       = "writing_sentence"
	UnitKeyWritingCharacter      = "writing_character"
	UnitKeySpeakingMinute        = "speaking_minute"
	UnitKeySpeakingDenseMinutes  = "speaking_dense_minutes"
	UnitKeyStudyMinute           = "study_minute"
)

type UnitDefinition struct {
	Key        string
	ActivityID int32
	Name       string
}

var unitDefinitions = []UnitDefinition{
	{Key: UnitKeyReadingPage, ActivityID: 1, Name: "Page"},
	{Key: UnitKeyReadingTwoColumnPage, ActivityID: 1, Name: "2 Column page"},
	{Key: UnitKeyReadingComicPage, ActivityID: 1, Name: "Comic page"},
	{Key: UnitKeyReadingSentence, ActivityID: 1, Name: "Sentence"},
	{Key: UnitKeyReadingCharacter, ActivityID: 1, Name: "Character"},
	{Key: UnitKeyListeningMinute, ActivityID: 2, Name: "Minute"},
	{Key: UnitKeyListeningDenseMinutes, ActivityID: 2, Name: "Minute (high density)"},
	{Key: UnitKeyWritingPage, ActivityID: 3, Name: "Page"},
	{Key: UnitKeyWritingSentence, ActivityID: 3, Name: "Sentence"},
	{Key: UnitKeyWritingCharacter, ActivityID: 3, Name: "Character"},
	{Key: UnitKeySpeakingMinute, ActivityID: 4, Name: "Minute"},
	{Key: UnitKeySpeakingDenseMinutes, ActivityID: 4, Name: "Minute (high density)"},
	{Key: UnitKeyStudyMinute, ActivityID: 5, Name: "Minute"},
}

var unitDefinitionsByKey = func() map[string]UnitDefinition {
	result := make(map[string]UnitDefinition, len(unitDefinitions))
	for _, definition := range unitDefinitions {
		result[definition.Key] = definition
	}
	return result
}()

func UnitDefinitions() []UnitDefinition {
	result := make([]UnitDefinition, len(unitDefinitions))
	copy(result, unitDefinitions)
	return result
}

func UnitDefinitionByKey(key string) (UnitDefinition, bool) {
	definition, ok := unitDefinitionsByKey[key]
	return definition, ok
}
