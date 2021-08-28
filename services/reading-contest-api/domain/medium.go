package domain

import (
	"github.com/srvc/fail"
)

// Medium knows how the score for a medium should be calculated
type Medium struct {
	Description string  `json:"description" db:"description"`
	Points      float32 `json:"points" db:"points"`
}

// Mediums is a collection of media
type Mediums map[MediumID]Medium

// MediumID is an id for media
type MediumID uint64

// These are the named constants for medium IDs
const (
	MediumBook MediumID = iota + 1
	MediumComic
	MediumNet
	MediumFullGame
	MediumGame
	MediumLyric
	MediumNews
	MediumSentences
)

// AllMediums is an array with all existing media
var AllMediums = Mediums{
	MediumBook:      Medium{Description: "Book", Points: 1},
	MediumComic:     Medium{Description: "Comic", Points: 0.2},
	MediumNet:       Medium{Description: "Net", Points: 1},
	MediumFullGame:  Medium{Description: "Full game", Points: 0.1667},
	MediumGame:      Medium{Description: "Game", Points: 0.05},
	MediumLyric:     Medium{Description: "Lyric", Points: 1},
	MediumNews:      Medium{Description: "News", Points: 1},
	MediumSentences: Medium{Description: "Sentences", Points: 0.05},
}

// ErrMediumNotFound for when a given medium id does not exist
var ErrMediumNotFound = fail.New("medium does not exist")

// Validate a MediumID
func (id MediumID) Validate() (bool, error) {
	found := false
	for currentID := range AllMediums {
		if currentID == id {
			found = true
		}
	}
	if !found {
		return false, ErrMediumNotFound
	}

	return true, nil
}

// AdjustedAmount gives the amount after having taken into account the medium
func (id MediumID) AdjustedAmount(amount float32) float32 {
	return AllMediums[id].Points * amount
}
