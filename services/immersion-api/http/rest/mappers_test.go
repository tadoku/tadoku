package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

func TestActivityToAPI(t *testing.T) {
	activity := domain.Activity{
		ID:        2,
		Name:      "Listening",
		Default:   true,
		InputType: domain.ActivityInputTypeTimePrimary,
	}

	res := activityToAPI(activity, true)

	require.NotNil(t, res.Default)
	require.NotNil(t, res.InputType)
	assert.Equal(t, int32(2), res.Id)
	assert.Equal(t, "Listening", res.Name)
	assert.True(t, *res.Default)
	assert.Equal(t, openapi.TimePrimary, *res.InputType)
}

func TestLogActivityToAPI(t *testing.T) {
	log := &domain.Log{
		ActivityID:   1,
		ActivityName: "Reading",
	}

	res := logActivityToAPI(log)

	require.NotNil(t, res.InputType)
	assert.Nil(t, res.Default)
	assert.Equal(t, int32(1), res.Id)
	assert.Equal(t, "Reading", res.Name)
	assert.Equal(t, openapi.AmountPrimary, *res.InputType)
}
