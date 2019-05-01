package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMedium_MediumIDValidate(t *testing.T) {
	{
		_, err := MediumID(20).Validate()
		assert.EqualError(t, err, ErrMediumNotFound.Error())
	}

	{
		_, err := MediumID(1).Validate()
		assert.NoError(t, err)
	}
}
