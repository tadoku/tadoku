package repository

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestTrackingNullConversions(t *testing.T) {
	unitID := uuid.New()

	tests := []struct {
		name     string
		tracking domain.LogTracking
	}{
		{
			name: "amount unit",
			tracking: domain.LogTracking{
				Kind:          domain.LogTrackingAmountUnit,
				UnitID:        unitID,
				Amount:        12.5,
				Modifier:      0.8,
				ComputedScore: 10,
			},
		},
		{
			name: "duration",
			tracking: domain.LogTracking{
				Kind:            domain.LogTrackingDuration,
				DurationSeconds: 600,
				ComputedScore:   5,
			},
		},
		{
			name: "both",
			tracking: domain.LogTracking{
				Kind:            domain.LogTrackingBoth,
				UnitID:          unitID,
				Amount:          12.5,
				Modifier:        0.8,
				DurationSeconds: 600,
				ComputedScore:   10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unit := trackingUnitID(tt.tracking)
			amount := trackingAmount(tt.tracking)
			modifier := trackingModifier(tt.tracking)
			duration := trackingDurationSeconds(tt.tracking)

			hasAmountUnit := tt.tracking.Kind == domain.LogTrackingAmountUnit || tt.tracking.Kind == domain.LogTrackingBoth
			assert.Equal(t, hasAmountUnit, unit.Valid)
			assert.Equal(t, hasAmountUnit, amount.Valid)
			assert.Equal(t, hasAmountUnit, modifier.Valid)
			if hasAmountUnit {
				assert.Equal(t, tt.tracking.UnitID, unit.UUID)
				assert.InDelta(t, tt.tracking.Amount, float32(amount.Float64), 0.0001)
				assert.InDelta(t, tt.tracking.Modifier, float32(modifier.Float64), 0.0001)
			}

			hasDuration := tt.tracking.Kind == domain.LogTrackingDuration || tt.tracking.Kind == domain.LogTrackingBoth
			assert.Equal(t, hasDuration, duration.Valid)
			if hasDuration {
				assert.Equal(t, tt.tracking.DurationSeconds, duration.Int32)
			}
		})
	}
}

func TestReadLogTracking(t *testing.T) {
	unitID := uuid.New()

	tests := []struct {
		name             string
		unitID           uuid.NullUUID
		amount           sql.NullFloat64
		modifier         sql.NullFloat64
		durationSeconds  sql.NullInt32
		score            sql.NullFloat64
		expectedTracking domain.LogTracking
	}{
		{
			name:     "amount unit",
			unitID:   uuid.NullUUID{UUID: unitID, Valid: true},
			amount:   sql.NullFloat64{Float64: 12.5, Valid: true},
			modifier: sql.NullFloat64{Float64: 0.8, Valid: true},
			score:    sql.NullFloat64{Float64: 10, Valid: true},
			expectedTracking: domain.LogTracking{
				Kind:          domain.LogTrackingAmountUnit,
				UnitID:        unitID,
				Amount:        12.5,
				Modifier:      0.8,
				ComputedScore: 10,
			},
		},
		{
			name:            "duration",
			durationSeconds: sql.NullInt32{Int32: 600, Valid: true},
			score:           sql.NullFloat64{Float64: 5, Valid: true},
			expectedTracking: domain.LogTracking{
				Kind:            domain.LogTrackingDuration,
				DurationSeconds: 600,
				ComputedScore:   5,
			},
		},
		{
			name:            "both",
			unitID:          uuid.NullUUID{UUID: unitID, Valid: true},
			amount:          sql.NullFloat64{Float64: 12.5, Valid: true},
			modifier:        sql.NullFloat64{Float64: 0.8, Valid: true},
			durationSeconds: sql.NullInt32{Int32: 600, Valid: true},
			score:           sql.NullFloat64{Float64: 10, Valid: true},
			expectedTracking: domain.LogTracking{
				Kind:            domain.LogTrackingBoth,
				UnitID:          unitID,
				Amount:          12.5,
				Modifier:        0.8,
				DurationSeconds: 600,
				ComputedScore:   10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracking := readLogTracking(tt.unitID, tt.amount, tt.modifier, tt.durationSeconds, tt.score)

			require.Equal(t, tt.expectedTracking.Kind, tracking.Kind)
			assert.Equal(t, tt.expectedTracking.UnitID, tracking.UnitID)
			assert.InDelta(t, tt.expectedTracking.Amount, tracking.Amount, 0.0001)
			assert.InDelta(t, tt.expectedTracking.Modifier, tracking.Modifier, 0.0001)
			assert.Equal(t, tt.expectedTracking.DurationSeconds, tracking.DurationSeconds)
			assert.InDelta(t, tt.expectedTracking.ComputedScore, tracking.ComputedScore, 0.0001)
		})
	}
}
