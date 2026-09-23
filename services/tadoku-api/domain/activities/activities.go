// Package activities defines the immersion activity catalog shared by logs and
// contests. It stays independent of both features so they can use the same
// definitions without importing each other.
package activities

type ActivityInputType string

const (
	ActivityInputTypeAmountPrimary ActivityInputType = "amount_primary"
	ActivityInputTypeTimePrimary   ActivityInputType = "time_primary"
)

var catalog = []Activity{
	{ID: 1, Name: "Reading", Default: true, InputType: ActivityInputTypeAmountPrimary, legacyDurationRate: .2},
	{ID: 2, Name: "Listening", Default: true, InputType: ActivityInputTypeTimePrimary, legacyDurationRate: .4},
	{ID: 3, Name: "Writing", Default: false, InputType: ActivityInputTypeAmountPrimary, legacyDurationRate: .2},
	{ID: 4, Name: "Speaking", Default: false, InputType: ActivityInputTypeTimePrimary, legacyDurationRate: .5},
	{ID: 5, Name: "Study", Default: false, InputType: ActivityInputTypeTimePrimary, legacyDurationRate: .5},
}

type Activity struct {
	ID                 int32
	Name               string
	Default            bool
	InputType          ActivityInputType
	legacyDurationRate float32
}

func All() []Activity { return append([]Activity(nil), catalog...) }

func LegacyDurationScorePerMinute(id int32) (float32, bool) {
	for _, activity := range catalog {
		if activity.ID == id {
			return activity.legacyDurationRate, true
		}
	}
	return 0, false
}

var unitActivities = map[string]int32{
	"reading_page":            1,
	"reading_two_column_page": 1,
	"reading_comic_page":      1,
	"reading_sentence":        1,
	"reading_character":       1,
	"listening_minute":        2,
	"listening_dense_minutes": 2,
	"writing_page":            3,
	"writing_sentence":        3,
	"writing_character":       3,
	"speaking_minute":         4,
	"speaking_dense_minutes":  4,
	"study_minute":            5,
}

func UnitActivityID(key string) (int32, bool) {
	id, ok := unitActivities[key]
	return id, ok
}
