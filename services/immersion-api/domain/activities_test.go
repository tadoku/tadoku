package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestActivities(t *testing.T) {
	t.Run("returns activities in id order", func(t *testing.T) {
		activities := domain.Activities()

		assert.Equal(t, []domain.Activity{
			{ID: 1, Name: "Reading", Default: true, InputType: domain.ActivityInputTypeAmountPrimary},
			{ID: 2, Name: "Listening", Default: true, InputType: domain.ActivityInputTypeTimePrimary},
			{ID: 3, Name: "Writing", Default: false, InputType: domain.ActivityInputTypeAmountPrimary},
			{ID: 4, Name: "Speaking", Default: false, InputType: domain.ActivityInputTypeTimePrimary},
			{ID: 5, Name: "Study", Default: false, InputType: domain.ActivityInputTypeTimePrimary},
		}, activities)
	})

	t.Run("looks up activities by id", func(t *testing.T) {
		activity, ok := domain.ActivityByID(1)

		assert.True(t, ok)
		assert.Equal(t, domain.Activity{
			ID:        1,
			Name:      "Reading",
			Default:   true,
			InputType: domain.ActivityInputTypeAmountPrimary,
		}, activity)
	})

	t.Run("returns false for unknown activity ids", func(t *testing.T) {
		_, ok := domain.ActivityByID(999)

		assert.False(t, ok)
		assert.False(t, domain.IsValidActivityID(999))
	})

	t.Run("sorts selected activities by name", func(t *testing.T) {
		activities, ok := domain.ActivitiesByIDsSortedByName([]int32{5, 1, 2})

		assert.True(t, ok)
		assert.Equal(t, []domain.Activity{
			{ID: 2, Name: "Listening", Default: true, InputType: domain.ActivityInputTypeTimePrimary},
			{ID: 1, Name: "Reading", Default: true, InputType: domain.ActivityInputTypeAmountPrimary},
			{ID: 5, Name: "Study", Default: false, InputType: domain.ActivityInputTypeTimePrimary},
		}, activities)
	})
}
