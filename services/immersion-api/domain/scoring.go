package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// validateTrackingData checks that the request has the right combination of
// time/amount based on the activity's input_type.
func validateTrackingData(activity *Activity, durationSeconds *int32, amount *float32, unitID *uuid.UUID) error {
	hasTime := durationSeconds != nil && *durationSeconds > 0
	hasAmount := amount != nil && unitID != nil && *amount > 0

	if activity.InputType == "time" {
		if !hasTime {
			return fmt.Errorf("time is required for this activity: %w", ErrInvalidLog)
		}
		return nil
	}

	// input_type == "amount": need at least time or amount
	if !hasTime && !hasAmount {
		return fmt.Errorf("time or amount is required: %w", ErrInvalidLog)
	}
	return nil
}

// ComputeScore calculates the score for a log based on the activity's input type.
// When amount is provided, it takes priority. When only time is provided, the
// activity's time_modifier is used.
func ComputeScore(activity *Activity, amount *float32, modifier *float32, durationSeconds *int32) float32 {
	hasAmount := amount != nil && modifier != nil

	if activity.InputType == "time" {
		// Time-based activities always score from time
		if durationSeconds != nil {
			return (float32(*durationSeconds) / 60.0) * activity.TimeModifier
		}
		return 0
	}

	// Amount-based activities: prefer amount, fall back to time
	if hasAmount {
		return *amount * *modifier
	}
	if durationSeconds != nil {
		return (float32(*durationSeconds) / 60.0) * activity.TimeModifier
	}
	return 0
}
