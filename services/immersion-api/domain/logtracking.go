package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

const (
	readingDurationScorePerMinute   = float32(0.2)
	listeningDurationScorePerMinute = float32(0.4)
	writingDurationScorePerMinute   = float32(0.2)
	speakingDurationScorePerMinute  = float32(0.5)
	studyDurationScorePerMinute     = float32(0.5)
)

type LogTrackingKind string

const (
	LogTrackingAmountUnit LogTrackingKind = "amount_unit"
	LogTrackingDuration   LogTrackingKind = "duration"
	LogTrackingBoth       LogTrackingKind = "both"
)

type LogTracking struct {
	Kind            LogTrackingKind
	UnitID          uuid.UUID
	UnitKey         string
	Amount          float32
	DurationSeconds int32
	Modifier        float32
	ComputedScore   float32
	ScoreProvenance *ScoreProvenance
}

type ScoreProvenance struct {
	RuleSetID *uuid.UUID
	RuleIDs   []uuid.UUID
	Rates     []float32
	Source    ScoreSource
}

type LogTrackingInput struct {
	ActivityID      int32
	UnitID          *uuid.UUID
	UnitKey         *string
	Amount          *float32
	Modifier        *float32
	DurationSeconds *int32
}

type UnitFindForTrackingRequest struct {
	ID           uuid.UUID
	ActivityID   int32
	LanguageCode string
}

type UnitFindForTrackingByKeyRequest struct {
	Key          string
	ActivityID   int32
	LanguageCode string
}

type logTrackingUnitFinder interface {
	FindUnitForTracking(context.Context, *UnitFindForTrackingRequest) (*Unit, error)
	FindUnitForTrackingByKey(context.Context, *UnitFindForTrackingByKeyRequest) (*Unit, error)
}

func DetermineLogTrackingKind(input LogTrackingInput) (LogTrackingKind, error) {
	hasAmount := input.Amount != nil
	hasUnit := input.UnitID != nil || input.UnitKey != nil
	hasDuration := input.DurationSeconds != nil

	if !IsValidActivityID(input.ActivityID) {
		return "", fmt.Errorf("activity %d is not valid: %w", input.ActivityID, ErrInvalidLog)
	}
	if hasDuration && *input.DurationSeconds <= 0 {
		return "", fmt.Errorf("duration_seconds must be positive: %w", ErrInvalidLog)
	}
	if hasAmount && *input.Amount <= 0 {
		return "", fmt.Errorf("amount must be positive: %w", ErrInvalidLog)
	}
	if hasAmount != hasUnit {
		return "", fmt.Errorf("amount and a unit identifier must be supplied together: %w", ErrInvalidLog)
	}
	if hasAmount && input.Modifier == nil {
		return "", fmt.Errorf("modifier is required for amount scoring: %w", ErrInvalidLog)
	}
	if !hasDuration && !hasAmount {
		return "", fmt.Errorf("amount/unit or duration_seconds is required: %w", ErrInvalidLog)
	}
	if hasAmount && hasDuration {
		return LogTrackingBoth, nil
	}
	if hasAmount {
		return LogTrackingAmountUnit, nil
	}
	return LogTrackingDuration, nil
}

func ValidateLogTracking(input LogTrackingInput) error {
	_, err := DetermineLogTrackingKind(input)
	return err
}

func ComputeInterimLogScore(input LogTrackingInput) (float32, error) {
	kind, err := DetermineLogTrackingKind(input)
	if err != nil {
		return 0, err
	}

	if kind == LogTrackingAmountUnit || kind == LogTrackingBoth {
		return *input.Amount * *input.Modifier, nil
	}

	minutes := float32(*input.DurationSeconds) / 60
	switch input.ActivityID {
	case 1:
		return minutes * readingDurationScorePerMinute, nil
	case 2:
		return minutes * listeningDurationScorePerMinute, nil
	case 3:
		return minutes * writingDurationScorePerMinute, nil
	case 4:
		return minutes * speakingDurationScorePerMinute, nil
	case 5:
		return minutes * studyDurationScorePerMinute, nil
	default:
		return 0, fmt.Errorf("activity %d is not valid: %w", input.ActivityID, ErrInvalidLog)
	}
}

func resolveLogTracking(
	ctx context.Context,
	finder logTrackingUnitFinder,
	activityID int32,
	languageCode string,
	unitID *uuid.UUID,
	unitKey *string,
	amount *float32,
	durationSeconds *int32,
) (LogTracking, error) {
	if !IsValidActivityID(activityID) {
		return LogTracking{}, fmt.Errorf("activity %d is not valid: %w", activityID, ErrInvalidLog)
	}

	var unit *Unit
	if unitKey != nil {
		definition, ok := UnitDefinitionByKey(*unitKey)
		if !ok || definition.ActivityID != activityID {
			return LogTracking{}, fmt.Errorf("unit key %q is not valid for activity %d: %w", *unitKey, activityID, ErrInvalidLog)
		}
	}
	if amount != nil && unitID != nil {
		resolved, err := finder.FindUnitForTracking(ctx, &UnitFindForTrackingRequest{
			ID:           *unitID,
			ActivityID:   activityID,
			LanguageCode: languageCode,
		})
		if err != nil {
			return LogTracking{}, err
		}
		unit = resolved
	}
	if amount != nil && unitID == nil && unitKey != nil {
		resolved, err := finder.FindUnitForTrackingByKey(ctx, &UnitFindForTrackingByKeyRequest{
			Key:          *unitKey,
			ActivityID:   activityID,
			LanguageCode: languageCode,
		})
		if err != nil {
			return LogTracking{}, err
		}
		unit = resolved
	}
	if unit != nil {
		definition, ok := UnitDefinitionByKey(unit.Key)
		if !ok || definition.ActivityID != activityID {
			return LogTracking{}, fmt.Errorf("resolved unit key %q is not valid for activity %d: %w", unit.Key, activityID, ErrInvalidLog)
		}
	}
	if unit != nil && unitKey != nil && unit.Key != *unitKey {
		return LogTracking{}, fmt.Errorf("unit_id and unit_key identify different units: %w", ErrInvalidLog)
	}

	var modifier *float32
	if unit != nil {
		modifier = &unit.Modifier
	}

	input := LogTrackingInput{
		ActivityID:      activityID,
		UnitID:          unitID,
		UnitKey:         unitKey,
		Amount:          amount,
		Modifier:        modifier,
		DurationSeconds: durationSeconds,
	}
	kind, err := DetermineLogTrackingKind(input)
	if err != nil {
		return LogTracking{}, err
	}
	computedScore, err := ComputeInterimLogScore(input)
	if err != nil {
		return LogTracking{}, err
	}

	tracking := LogTracking{
		Kind:          kind,
		ComputedScore: computedScore,
	}
	if kind == LogTrackingAmountUnit || kind == LogTrackingBoth {
		tracking.UnitID = unit.ID
		tracking.UnitKey = unit.Key
		tracking.Amount = *amount
		tracking.Modifier = *modifier
	}
	if kind == LogTrackingDuration || kind == LogTrackingBoth {
		tracking.DurationSeconds = *durationSeconds
	}

	return tracking, nil
}
