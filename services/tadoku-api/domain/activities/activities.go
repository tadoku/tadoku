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
	{ID: 1, Name: "Reading", Default: true, InputType: ActivityInputTypeAmountPrimary},
	{ID: 2, Name: "Listening", Default: true, InputType: ActivityInputTypeTimePrimary},
	{ID: 3, Name: "Writing", Default: false, InputType: ActivityInputTypeAmountPrimary},
	{ID: 4, Name: "Speaking", Default: false, InputType: ActivityInputTypeTimePrimary},
	{ID: 5, Name: "Study", Default: false, InputType: ActivityInputTypeTimePrimary},
}

type Activity struct {
	ID        int32
	Name      string
	Default   bool
	InputType ActivityInputType
}

func All() []Activity { return append([]Activity(nil), catalog...) }

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
