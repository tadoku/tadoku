package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMedium_MediumIDValidate(t *testing.T) {
	{
		_, err := MediumID(20).Validate()
		assert.Equal(t, err, ErrMediumNotFound)
	}

	{
		_, err := MediumID(1).Validate()
		assert.Nil(t, err)
	}
}
