package scoring

import (
	"context"
	"math"
	"slices"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

const (
	readingActivityID = 1
	pageUnitKey       = "pages"
)

var minuteUnitTags = map[string]string{
	"listening_minute":        "",
	"listening_dense_minutes": "dense",
	"speaking_minute":         "",
	"speaking_dense_minutes":  "dense",
	"study_minute":            "",
}

var retiredUnitTags = map[string]string{
	"reading_comic_page":      "comic",
	"reading_two_column_page": "two_column",
}

// normalizeLegacyUnit rewrites an amount in a minute or retired unit into the shape the platform rule set scores, and
// counts that and any activity-prefixed key spelling. Input it cannot resolve is returned unchanged, so the caller's
// validation reports it.
func (s *Service) normalizeLegacyUnit(ctx context.Context, input logscore.Input) (logscore.Input, error) {
	if input.Amount == nil || (input.UnitID == nil && input.UnitKey == nil) {
		return input, nil
	}

	unit, err := s.findApplicableUnit(ctx, input.UnitID, input.UnitKey, input.ActivityID, input.LanguageCode)
	if err != nil || unit == nil {
		return input, err
	}

	var page *logUnit
	if _, retired := retiredUnitTags[unit.Key]; retired {
		key := pageUnitKey
		page, err = s.findApplicableUnit(ctx, nil, &key, readingActivityID, "")
		if err != nil {
			return input, err
		}
	}

	normalized, kind, err := convertLegacyUnit(input, *unit, page)
	if err != nil {
		return input, err
	}
	if kind != "" {
		s.observer.ObserveUnitNormalization(ctx, kind)
	}
	return normalized, nil
}

func convertLegacyUnit(input logscore.Input, unit logUnit, page *logUnit) (logscore.Input, string, error) {
	if input.UnitKey != nil && activities.CanonicalUnitKey(input.ActivityID, *input.UnitKey) != unit.Key {
		return input, "", nil
	}
	if tag, minutes := minuteUnitTags[unit.Key]; minutes {
		return convertMinuteAmount(input, tag)
	}

	if tag, retired := retiredUnitTags[unit.Key]; retired {
		if page == nil {
			return input, "", errx.NewInternalError("reading page log unit is missing")
		}
		tags, err := normalizeTags(append(slices.Clone(input.Tags), tag))
		if err != nil {
			return input, "", err
		}

		input.UnitID = &page.ID
		input.UnitKey = &page.Key
		input.Tags = tags
		return input, "retired_unit", nil
	}

	if input.UnitKey != nil && activities.IsLegacyUnitKey(*input.UnitKey) {
		return input, "legacy_key_spelling", nil
	}
	return input, "", nil
}

func convertMinuteAmount(input logscore.Input, tag string) (logscore.Input, string, error) {
	amount := *input.Amount
	if !finite(amount) || amount <= 0 || (input.DurationSeconds != nil && *input.DurationSeconds <= 0) {
		return input, "", nil
	}

	if input.DurationSeconds == nil {
		seconds := math.Round(float64(amount) * 60)
		if seconds < 1 {
			return input, "", errx.NewInvalidInputError("amount in minutes must be at least one second")
		}
		if seconds > math.MaxInt32 {
			return input, "", errx.NewInvalidInputError("amount in minutes exceeds the duration_seconds range")
		}
		duration := int32(seconds)
		input.DurationSeconds = &duration
	}
	if tag != "" {
		tags, err := normalizeTags(append(slices.Clone(input.Tags), tag))
		if err != nil {
			return input, "", err
		}
		input.Tags = tags
	}

	input.Amount = nil
	input.UnitID = nil
	input.UnitKey = nil
	return input, "minute_amount", nil
}
