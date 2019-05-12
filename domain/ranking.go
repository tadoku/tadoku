package domain

import (
	"time"
)

// Ranking contains the data about a user that has entered a contest
type Ranking struct {
	ID        uint64       `json:"id" db:"id"`
	ContestID uint64       `json:"contest_id" db:"contest_id" valid:"required"`
	UserID    uint64       `json:"user_id" db:"user_id" valid:"required"`
	Language  LanguageCode `json:"language_code" db:"language_code" valid:"required"`
	Amount    float32      `json:"amount" db:"amount" valid:"required"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`

	// Optional fields
	UserDisplayName string `json:"user_display_name" db:"user_display_name"`
}

// GetView gets the external view representation of a Ranking
func (r Ranking) GetView() RankingView {
	return RankingView{
		UserID:          r.UserID,
		UserDisplayName: r.UserDisplayName,
		Language:        r.Language,
		Amount:          r.Amount,
	}
}

// Rankings is a collection of Ranking
type Rankings []Ranking

// GetView gets the external view representation of a Rankings collection
func (r Rankings) GetView() []RankingView {
	result := make([]RankingView, len(r))

	for i, val := range r {
		result[i] = val.GetView()
	}

	return result
}

// RankingView is a representation of a ranking for external usages
type RankingView struct {
	UserID          uint64       `json:"user_id"`
	UserDisplayName string       `json:"user_display_name"`
	Language        LanguageCode `json:"language_code"`
	Amount          float32      `json:"amount"`
}
