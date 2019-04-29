package domain

import (
	"time"
)

// ContestLog contains a single entry in a contest for a user
type ContestLog struct {
	ID        uint64       `json:"id" db:"id"`
	ContestID uint64       `json:"contest_id" db:"contest_id" valid:"required"`
	UserID    uint64       `json:"user_id" db:"user_id" valid:"required"`
	Language  LanguageCode `json:"language_code" db:"language_code" valid:"required"`
	Amount    float32      `json:"amount" db:"amount" valid:"required"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at" db:"deleted_at"`
}

// ContestLogs is a collection of ContestLog
type ContestLogs []ContestLog
