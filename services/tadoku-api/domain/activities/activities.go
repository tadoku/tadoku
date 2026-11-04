package activities

import "slices"

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

var unitActivities = map[string][]int32{
	"pages":                   {1, 3},
	"sentences":               {1, 3},
	"characters":              {1, 3},
	"reading_page":            {1},
	"reading_two_column_page": {1},
	"reading_comic_page":      {1},
	"reading_sentence":        {1},
	"reading_character":       {1},
	"listening_minute":        {2},
	"listening_dense_minutes": {2},
	"writing_page":            {3},
	"writing_sentence":        {3},
	"writing_character":       {3},
	"speaking_minute":         {4},
	"speaking_dense_minutes":  {4},
	"study_minute":            {5},
}

var legacyUnitKeys = map[string]string{
	"reading_page":      "pages",
	"reading_sentence":  "sentences",
	"reading_character": "characters",
	"writing_page":      "pages",
	"writing_sentence":  "sentences",
	"writing_character": "characters",
}

func UnitValidForActivity(key string, activityID int32) bool {
	return slices.Contains(unitActivities[key], activityID)
}

// CanonicalUnitKey returns the plural key for an activity-prefixed reading or writing key of the activity, and key
// otherwise. Compare unit keys in this form: stored log units and rules may carry either spelling.
func CanonicalUnitKey(activityID int32, key string) string {
	if plural, legacy := legacyUnitKeys[key]; legacy && UnitValidForActivity(key, activityID) {
		return plural
	}
	return key
}

// UnitKeySpellings returns key followed by its other spelling for the activity, when it has one.
func UnitKeySpellings(activityID int32, key string) []string {
	if plural, legacy := legacyUnitKeys[key]; legacy && UnitValidForActivity(key, activityID) {
		return []string{key, plural}
	}
	for legacy, plural := range legacyUnitKeys {
		if plural == key && UnitValidForActivity(legacy, activityID) {
			return []string{key, legacy}
		}
	}
	return []string{key}
}

var pluralUnitKeys = []string{"pages", "sentences", "characters"}

func PluralUnitKeys() []string { return slices.Clone(pluralUnitKeys) }

func IsPluralUnitKey(key string) bool { return slices.Contains(pluralUnitKeys, key) }

func IsLegacyUnitKey(key string) bool {
	_, legacy := legacyUnitKeys[key]
	return legacy
}
