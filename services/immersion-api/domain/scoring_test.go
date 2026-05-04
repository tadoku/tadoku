package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T { return &v }

func TestValidateTrackingData(t *testing.T) {
	timeActivity := &Activity{ID: 1, Name: "Listening", InputType: "time"}
	amountActivity := &Activity{ID: 2, Name: "Reading", InputType: "amount"}

	t.Run("time activity: valid with duration", func(t *testing.T) {
		err := validateTrackingData(timeActivity, ptr(int32(3600)), nil, nil)
		assert.NoError(t, err)
	})

	t.Run("time activity: error without duration", func(t *testing.T) {
		err := validateTrackingData(timeActivity, nil, nil, nil)
		assert.ErrorIs(t, err, ErrInvalidLog)
	})

	t.Run("time activity: error with zero duration", func(t *testing.T) {
		err := validateTrackingData(timeActivity, ptr(int32(0)), nil, nil)
		assert.ErrorIs(t, err, ErrInvalidLog)
	})

	t.Run("time activity: ignores amount even if provided", func(t *testing.T) {
		unitID := uuid.New()
		err := validateTrackingData(timeActivity, ptr(int32(3600)), ptr(float32(10)), &unitID)
		assert.NoError(t, err)
	})

	t.Run("amount activity: valid with amount and unit", func(t *testing.T) {
		unitID := uuid.New()
		err := validateTrackingData(amountActivity, nil, ptr(float32(10)), &unitID)
		assert.NoError(t, err)
	})

	t.Run("amount activity: valid with only duration", func(t *testing.T) {
		err := validateTrackingData(amountActivity, ptr(int32(3600)), nil, nil)
		assert.NoError(t, err)
	})

	t.Run("amount activity: valid with both amount and duration", func(t *testing.T) {
		unitID := uuid.New()
		err := validateTrackingData(amountActivity, ptr(int32(3600)), ptr(float32(10)), &unitID)
		assert.NoError(t, err)
	})

	t.Run("amount activity: error with neither", func(t *testing.T) {
		err := validateTrackingData(amountActivity, nil, nil, nil)
		assert.ErrorIs(t, err, ErrInvalidLog)
	})

	t.Run("amount activity: error with amount but no unit", func(t *testing.T) {
		err := validateTrackingData(amountActivity, nil, ptr(float32(10)), nil)
		assert.ErrorIs(t, err, ErrInvalidLog)
	})

	t.Run("amount activity: error with unit but no amount", func(t *testing.T) {
		unitID := uuid.New()
		err := validateTrackingData(amountActivity, nil, nil, &unitID)
		assert.ErrorIs(t, err, ErrInvalidLog)
	})

	t.Run("amount activity: error with zero amount", func(t *testing.T) {
		unitID := uuid.New()
		err := validateTrackingData(amountActivity, nil, ptr(float32(0)), &unitID)
		assert.ErrorIs(t, err, ErrInvalidLog)
	})
}

func TestComputeScore(t *testing.T) {
	timeActivity := &Activity{ID: 1, Name: "Listening", InputType: "time"}
	amountActivity := &Activity{ID: 2, Name: "Reading", InputType: "amount"}

	t.Run("time activity: scores from duration", func(t *testing.T) {
		// 3600 seconds = 60 minutes, 60 * 0.3 = 18
		score := ComputeScore(timeActivity, nil, nil, ptr(int32(3600)))
		assert.InDelta(t, float32(18.0), score, 0.01)
	})

	t.Run("time activity: ignores amount even if provided", func(t *testing.T) {
		// Should use time, not amount
		score := ComputeScore(timeActivity, ptr(float32(100)), ptr(float32(1.0)), ptr(int32(3600)))
		assert.InDelta(t, float32(18.0), score, 0.01)
	})

	t.Run("time activity: returns 0 without duration", func(t *testing.T) {
		score := ComputeScore(timeActivity, nil, nil, nil)
		assert.Equal(t, float32(0), score)
	})

	t.Run("amount activity: scores from amount and modifier", func(t *testing.T) {
		// 10 * 1.5 = 15
		score := ComputeScore(amountActivity, ptr(float32(10)), ptr(float32(1.5)), nil)
		assert.InDelta(t, float32(15.0), score, 0.01)
	})

	t.Run("amount activity: prefers amount over time", func(t *testing.T) {
		// Should use amount (10 * 1.5 = 15), not time (60 * 0.3 = 18)
		score := ComputeScore(amountActivity, ptr(float32(10)), ptr(float32(1.5)), ptr(int32(3600)))
		assert.InDelta(t, float32(15.0), score, 0.01)
	})

	t.Run("amount activity: falls back to time when no amount", func(t *testing.T) {
		// 3600 seconds = 60 minutes, 60 * 0.3 = 18
		score := ComputeScore(amountActivity, nil, nil, ptr(int32(3600)))
		assert.InDelta(t, float32(18.0), score, 0.01)
	})

	t.Run("amount activity: returns 0 with neither", func(t *testing.T) {
		score := ComputeScore(amountActivity, nil, nil, nil)
		assert.Equal(t, float32(0), score)
	})

	t.Run("amount activity: falls back to time when amount present but no modifier", func(t *testing.T) {
		score := ComputeScore(amountActivity, ptr(float32(10)), nil, ptr(int32(3600)))
		assert.InDelta(t, float32(18.0), score, 0.01)
	})
}
