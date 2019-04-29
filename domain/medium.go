package domain

// Medium knows how the score for a medium should be calculated
type Medium struct {
	Description string  `json:"description" db:"description"`
	Points      float32 `json:"points" db:"points"`
}

// Mediums is a collection of media
type Mediums map[uint64]Medium

// AllMediums is an array with all existing media
var AllMediums = Mediums{
	1: Medium{Description: "Book", Points: 1},
	2: Medium{Description: "Manga", Points: 0.2},
	3: Medium{Description: "Net", Points: 1},
	4: Medium{Description: "Full game", Points: 0.1667},
	5: Medium{Description: "Game", Points: 0.05},
	6: Medium{Description: "Lyric", Points: 1},
	7: Medium{Description: "Subs", Points: 0.2},
	8: Medium{Description: "News", Points: 1},
	9: Medium{Description: "Sentences", Points: 0.05},
}
