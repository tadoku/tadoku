package logscore

import "github.com/google/uuid"

type Input struct {
	UnitID          *uuid.UUID
	UnitKey         *string
	ActivityID      int32
	LanguageCode    string
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
}

type Target struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
	Official       bool
}

type Tracking struct {
	UnitID          *uuid.UUID
	UnitKey         string
	Amount          *float32
	Modifier        *float32
	DurationSeconds *int32
	Score           float32
	RuleSetID       *uuid.UUID
	RuleIDs         []uuid.UUID
	Rates           []float32
	Source          string
}

type ContestTracking struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
	Tracking       Tracking
}

type Result struct {
	ActivityID       int32
	LanguageCode     string
	Tags             []string
	Tracking         Tracking
	ContestTrackings []ContestTracking
	EligibleOfficial bool
}
