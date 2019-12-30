package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_PasswordIsNotExported(t *testing.T) {
	u := User{
		ID:          1,
		Email:       "foo@example.com",
		DisplayName: "John Smith",
		Password:    "foobar",
		Role:        RoleUser,
		Preferences: &Preferences{},
	}
	userJSON, err := json.Marshal(u)
	assert.NoError(t, err)

	newUser := User{}
	json.Unmarshal(userJSON, &newUser)
	assert.Empty(t, newUser.Password)
}

func TestUser_Validation(t *testing.T) {
	var tests = []struct {
		user          User
		expectedError error
	}{
		{User{Password: "ewgflikhghewioghew"}, nil},
		{User{}, ErrUserMissingPassword},
	}

	for _, test := range tests {
		_, err := test.user.Validate()

		if test.expectedError != nil {
			assert.EqualErrorf(t, err, test.expectedError.Error(), "expected user.Validate of %v to be %v, got %v instead", test.user, test.expectedError, err)
		} else {
			assert.NoErrorf(t, err, "expected user.Validate of %v to be %v, got %v instead", test.user, test.expectedError, err)
		}
	}
}
