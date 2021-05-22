package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContestLog_Validate(t *testing.T) {
	// With existing medium
	{
		log := ContestLog{
			ContestID: 1,
			UserID:    1,
			Language:  Japanese,
			Amount:    10,
			MediumID:  1,
		}

		valid, err := log.Validate()
		assert.Equal(t, true, valid)
		assert.NoError(t, err)
	}

	// With invalid medium
	{
		log := ContestLog{
			ContestID: 1,
			UserID:    1,
			Language:  Japanese,
			Amount:    10,
			MediumID:  20,
		}

		valid, err := log.Validate()
		assert.Equal(t, false, valid)
		assert.Equal(t, ErrMediumNotFound, err)
	}
}
