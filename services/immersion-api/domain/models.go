package domain

import (
	"time"

	"github.com/google/uuid"
)

type Language struct {
	Code string
	Name string
}

type Activity struct {
	ID      int32
	Name    string
	Default bool
}

type ContestView struct {
	ID                   uuid.UUID
	ContestStart         time.Time
	ContestEnd           time.Time
	RegistrationEnd      time.Time
	Title                string
	Description          *string
	OwnerUserID          uuid.UUID
	OwnerUserDisplayName string
	Official             bool
	Private              bool
	AllowedLanguages     []Language
	AllowedActivities    []Activity
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Deleted              bool
}

type Contest struct {
	ID                      uuid.UUID
	ContestStart            time.Time
	ContestEnd              time.Time
	RegistrationEnd         time.Time
	Title                   string
	Description             *string
	OwnerUserID             uuid.UUID
	OwnerUserDisplayName    string
	Official                bool
	Private                 bool
	LanguageCodeAllowList   []string
	ActivityTypeIDAllowList []int32
	CreatedAt               time.Time
	UpdatedAt               time.Time
	Deleted                 bool
}

type ContestRegistration struct {
	ID              uuid.UUID
	ContestID       uuid.UUID
	UserID          uuid.UUID
	UserDisplayName string
	Languages       []Language
	Contest         *ContestView
}

type ContestRegistrationReference struct {
	RegistrationID       uuid.UUID
	ContestID            uuid.UUID
	ContestEnd           time.Time
	Title                string
	OwnerUserDisplayName string
	Official             bool
	Score                float32
}

type ContestRegistrations struct {
	Registrations []ContestRegistration
	TotalSize     int
	NextPageToken string
}

type Leaderboard struct {
	Entries       []LeaderboardEntry
	TotalSize     int
	NextPageToken string
}

type LeaderboardEntry struct {
	Rank            int
	UserID          uuid.UUID
	UserDisplayName string
	Score           float32
	IsTie           bool
}

type Score struct {
	LanguageCode string
	LanguageName *string
	Score        float32
}

type UserTraits struct {
	UserDisplayName string
	Email           string
	CreatedAt       time.Time
}

type UserProfile struct {
	DisplayName string
	CreatedAt   time.Time
}

type UserActivityScore struct {
	Date    time.Time
	Score   float32
	Updates int
}

type ActivityScore struct {
	ActivityID   int
	ActivityName string
	Score        float32
}

type Unit struct {
	ID            uuid.UUID
	LogActivityID int
	Name          string
	Modifier      float32
	LanguageCode  *string
}

type Tag struct {
	ID   uuid.UUID
	Name string
}

type Log struct {
	ID                          uuid.UUID
	UserID                      uuid.UUID
	UserDisplayName             *string
	Description                 *string
	LanguageCode                string
	LanguageName                string
	ActivityID                  int
	ActivityName                string
	UnitName                    string
	Tags                        []string
	Amount                      float32
	Modifier                    float32
	Score                       float32
	EligibleOfficialLeaderboard bool
	Registrations               []ContestRegistrationReference
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	Deleted                     bool
}
