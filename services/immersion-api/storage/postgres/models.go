// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Contest struct {
	ID                      uuid.UUID
	OwnerUserID             uuid.NullUUID
	OwnerUserDisplayName    sql.NullString
	Private                 bool
	ContestStart            time.Time
	ContestEnd              time.Time
	RegistrationStart       time.Time
	RegistrationEnd         time.Time
	Description             string
	LanguageCodeAllowList   []string
	ActivityTypeIDAllowList []int16
	Official                sql.NullBool
	CreatedAt               time.Time
	UpdatedAt               time.Time
	DeletedAt               sql.NullTime
}

type ContestRegistration struct {
	ID              uuid.UUID
	ContestID       uuid.UUID
	UserID          uuid.UUID
	UserDisplayName string
	LanguageCodes   []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       sql.NullTime
}

type Language struct {
	// See https://en.wikipedia.org/wiki/Wikipedia:WikiProject_Languages/List_of_ISO_639-3_language_codes_(2019)
	Code string
	Name string
}

type Log struct {
	ID                          uuid.UUID
	GroupID                     uuid.UUID
	ContestID                   uuid.NullUUID
	UserID                      uuid.UUID
	LanguageCode                string
	LogActivityTypeID           int16
	UnitID                      uuid.UUID
	Tags                        []string
	Amount                      float32
	Modifier                    float32
	Score                       float32
	EligibleOfficialLeaderboard bool
	Year                        int16
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	DeletedAt                   sql.NullTime
}

type LogActivityType struct {
	ID   int16
	Name string
}

type LogTag struct {
	ID                uuid.UUID
	LogActivityTypeID int16
	Name              string
}

type LogUnit struct {
	ID                uuid.UUID
	LogActivityTypeID int16
	Unit              string
	Modifier          float32
	LanguageCode      sql.NullString
}
