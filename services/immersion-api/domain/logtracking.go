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
	Amount          float32
	DurationSeconds int32
	Modifier        float32
	ComputedScore   float32
}

type LogTrackingInput struct {
	ActivityID      int32
	UnitID          *uuid.UUID
	Amount          *float32
	Modifier        *float32
	DurationSeconds *int32
}

type UnitFindForTrackingRequest struct {
	ID           uuid.UUID
	ActivityID   int32
	LanguageCode string
}

type logTrackingUnitFinder interface {
	FindUnitForTracking(context.Context, *UnitFindForTrackingRequest) (*Unit, error)
}

func DetermineLogTrackingKind(input LogTrackingInput) (LogTrackingKind, error) {
	hasAmount := input.Amount != nil
	hasUnit := input.UnitID != nil
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
		return "", fmt.Errorf("amount and unit_id must be supplied together: %w", ErrInvalidLog)
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
	amount *float32,
	durationSeconds *int32,
) (LogTracking, error) {
	if !IsValidActivityID(activityID) {
		return LogTracking{}, fmt.Errorf("activity %d is not valid: %w", activityID, ErrInvalidLog)
	}

	var modifier *float32
	if amount != nil && unitID != nil {
		unit, err := finder.FindUnitForTracking(ctx, &UnitFindForTrackingRequest{
			ID:           *unitID,
			ActivityID:   activityID,
			LanguageCode: languageCode,
		})
		if err != nil {
			return LogTracking{}, err
		}
		modifier = &unit.Modifier
	}

	input := LogTrackingInput{
		ActivityID:      activityID,
		UnitID:          unitID,
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
		tracking.UnitID = *unitID
		tracking.Amount = *amount
		tracking.Modifier = *modifier
	}
	if kind == LogTrackingDuration || kind == LogTrackingBoth {
		tracking.DurationSeconds = *durationSeconds
	}

	return tracking, nil
}
