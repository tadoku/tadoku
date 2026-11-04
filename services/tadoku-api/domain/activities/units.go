package activities

import (
	"errors"
	"maps"
	"slices"
)

var (
	errUnknownUnit    = errors.New("unit_key is not a known unit")
	errUnitNotOffered = errors.New("unit_key is not offered by activity_id")
)

type Unit struct {
	ActivityID   int32
	Key          string
	Name         string
	Modifier     float32
	LanguageCode *string
}

type catalogUnit struct {
	activityID        int32
	key               string
	name              string
	modifier          float32
	languageModifiers map[string]float32
}

var unitCatalog = []catalogUnit{
	{activityID: 1, key: "pages", name: "Page", modifier: 1},
	{activityID: 1, key: "sentences", name: "Sentence", modifier: 0.05},
	{
		activityID:        1,
		key:               "characters",
		name:              "Character",
		modifier:          0.000833333,
		languageModifiers: logographicLanguages(0.0025),
	},
	{activityID: 3, key: "pages", name: "Page", modifier: 10},
	{activityID: 3, key: "sentences", name: "Sentence", modifier: 0.5},
	{
		activityID:        3,
		key:               "characters",
		name:              "Character",
		modifier:          0.00833333,
		languageModifiers: logographicLanguages(0.025),
	},
}

func logographicLanguages(modifier float32) map[string]float32 {
	modifiers := make(map[string]float32)
	for _, code := range []string{"cmn", "jpn", "kor", "wuu", "yue", "zho"} {
		modifiers[code] = modifier
	}
	return modifiers
}

var storedUnitKeys = map[string]catalogUnit{
	"reading_page":      {activityID: 1, key: "pages"},
	"reading_sentence":  {activityID: 1, key: "sentences"},
	"reading_character": {activityID: 1, key: "characters"},
	"writing_page":      {activityID: 3, key: "pages"},
	"writing_sentence":  {activityID: 3, key: "sentences"},
	"writing_character": {activityID: 3, key: "characters"},
}

// Units lists every unit an activity offers, by activity, then pages, sentences and characters, each with its default
// modifier before its language modifiers in language code order.
func Units() []Unit {
	var units []Unit
	for _, unit := range unitCatalog {
		units = append(units, unit.resolve(nil))
		for _, code := range slices.Sorted(maps.Keys(unit.languageModifiers)) {
			units = append(units, unit.resolve(&code))
		}
	}
	return units
}

// ResolveUnit returns the unit of activityID named key, with the modifier for languageCode. Its error message names
// the unit_key and activity_id fields, so callers can report it as invalid input.
func ResolveUnit(activityID int32, key, languageCode string) (Unit, error) {
	known := false
	for _, unit := range unitCatalog {
		if unit.key != key {
			continue
		}
		known = true
		if unit.activityID != activityID {
			continue
		}

		if _, ok := unit.languageModifiers[languageCode]; ok {
			return unit.resolve(&languageCode), nil
		}
		return unit.resolve(nil), nil
	}

	if !known {
		return Unit{}, errUnknownUnit
	}
	return Unit{}, errUnitNotOffered
}

func (u catalogUnit) resolve(languageCode *string) Unit {
	unit := Unit{
		ActivityID: u.activityID,
		Key:        u.key,
		Name:       u.name,
		Modifier:   u.modifier,
	}
	if languageCode != nil {
		unit.Modifier = u.languageModifiers[*languageCode]
		unit.LanguageCode = languageCode
	}
	return unit
}

// CanonicalStoredUnitKey maps an activity-prefixed page, sentence or character key that a stored scoring rule may still
// hold to its catalog key, and returns every other key unchanged. Never use it on request input.
func CanonicalStoredUnitKey(activityID int32, key string) string {
	if unit, stored := storedUnitKeys[key]; stored && unit.activityID == activityID {
		return unit.key
	}
	return key
}
