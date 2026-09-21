// Package logs owns immersion log configuration data and tag suggestions.
package logs

import "github.com/google/uuid"

type Unit struct {
	ID            uuid.UUID
	Key           string
	LogActivityID int
	Name          string
	Modifier      float32
	LanguageCode  *string
}

type ConfigurationOptions struct {
	Units                []Unit
	UserLanguageCodes    []string
	ScoringEngineEnabled bool
}

type TagSuggestion struct {
	Tag   string
	Count int
}
