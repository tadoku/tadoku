package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestComputeInterimLogScore(t *testing.T) {
	t.Run("scores amount and unit tracking from amount and modifier", func(t *testing.T) {
		unitID := uuid.New()
		amount := float32(10)
		modifier := float32(1.6)
		durationSeconds := int32(999)

		score, err := domain.ComputeInterimLogScore(domain.LogTrackingInput{
			ActivityID:      1,
			UnitID:          &unitID,
			Amount:          &amount,
			Modifier:        &modifier,
			DurationSeconds: &durationSeconds,
		})

		require.NoError(t, err)
		assert.InDelta(t, float32(16), score, 0.0001)
	})

	t.Run("scores duration-only listening from the manual plain minute row", func(t *testing.T) {
		durationSeconds := int32(120)

		score, err := domain.ComputeInterimLogScore(domain.LogTrackingInput{
			ActivityID:      2,
			DurationSeconds: &durationSeconds,
		})

		require.NoError(t, err)
		assert.InDelta(t, float32(0.8), score, 0.0001)
	})

	t.Run("scores duration-only speaking and study from the manual plain minute row", func(t *testing.T) {
		durationSeconds := int32(120)

		for _, activityID := range []int32{4, 5} {
			score, err := domain.ComputeInterimLogScore(domain.LogTrackingInput{
				ActivityID:      activityID,
				DurationSeconds: &durationSeconds,
			})

			require.NoError(t, err)
			assert.InDelta(t, float32(1), score, 0.0001)
		}
	})

	t.Run("scores duration-only reading and writing at one fifth point per minute", func(t *testing.T) {
		durationSeconds := int32(300)

		for _, activityID := range []int32{1, 3} {
			score, err := domain.ComputeInterimLogScore(domain.LogTrackingInput{
				ActivityID:      activityID,
				DurationSeconds: &durationSeconds,
			})

			require.NoError(t, err)
			assert.InDelta(t, float32(1), score, 0.0001)
		}
	})

	t.Run("rejects invalid tracking combinations", func(t *testing.T) {
		unitID := uuid.New()
		amount := float32(10)
		zeroDuration := int32(0)
		negativeAmount := float32(-1)
		amountWithoutModifier := float32(1)
		unknownActivityDuration := int32(60)

		tests := []struct {
			name  string
			input domain.LogTrackingInput
		}{
			{
				name: "no tracking",
				input: domain.LogTrackingInput{
					ActivityID: 1,
				},
			},
			{
				name: "only amount",
				input: domain.LogTrackingInput{
					ActivityID: 1,
					Amount:     &amount,
				},
			},
			{
				name: "only unit",
				input: domain.LogTrackingInput{
					ActivityID: 1,
					UnitID:     &unitID,
				},
			},
			{
				name: "zero duration",
				input: domain.LogTrackingInput{
					ActivityID:      1,
					DurationSeconds: &zeroDuration,
				},
			},
			{
				name: "negative amount",
				input: domain.LogTrackingInput{
					ActivityID: 1,
					UnitID:     &unitID,
					Amount:     &negativeAmount,
				},
			},
			{
				name: "missing modifier",
				input: domain.LogTrackingInput{
					ActivityID: 1,
					UnitID:     &unitID,
					Amount:     &amountWithoutModifier,
				},
			},
			{
				name: "unknown activity",
				input: domain.LogTrackingInput{
					ActivityID:      999,
					DurationSeconds: &unknownActivityDuration,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := domain.ComputeInterimLogScore(tt.input)

				assert.ErrorIs(t, err, domain.ErrInvalidLog)
			})
		}
	})
}
