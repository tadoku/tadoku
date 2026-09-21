// Package logs owns immersion log data, derived statistics and configuration.
package logs

import (
	"github.com/google/uuid"
	"time"
)

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

type ActivityScore struct {
	Date    time.Time
	Score   float32
	Updates int
}

type YearlyActivity struct {
	Scores       []ActivityScore
	TotalUpdates int
}

type Score struct {
	LanguageCode string
	LanguageName string
	Score        float32
}

type YearlyScores struct {
	Scores       []Score
	OverallScore float32
}

type ActivitySplitScore struct {
	ActivityID   int
	ActivityName string
	Score        float32
}
