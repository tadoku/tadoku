package domain

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
	MediumBook MediumID = iota
	MediumManga
	MediumNet
	MediumFullGame
	MediumGame
	MediumLyric
	MediumSubs
	MediumNews
	MediumSentences
)

// AllMediums is an array with all existing media
var AllMediums = Mediums{
	MediumBook:      Medium{Description: "Book", Points: 1},
	MediumManga:     Medium{Description: "Manga", Points: 0.2},
	MediumNet:       Medium{Description: "Net", Points: 1},
	MediumFullGame:  Medium{Description: "Full game", Points: 0.1667},
	MediumGame:      Medium{Description: "Game", Points: 0.05},
	MediumLyric:     Medium{Description: "Lyric", Points: 1},
	MediumSubs:      Medium{Description: "Subs", Points: 0.2},
	MediumNews:      Medium{Description: "News", Points: 1},
	MediumSentences: Medium{Description: "Sentences", Points: 0.05},
}
