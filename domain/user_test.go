package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
)

func TestUser_PasswordIsNotExported(t *testing.T) {
	u := domain.User{
		ID:          1,
		Email:       "foo@example.com",
		DisplayName: "John Smith",
		Password:    "foobar",
		Role:        domain.RoleUser,
		Preferences: &domain.Preferences{},
	}
	userJSON, err := json.Marshal(u)
	assert.NoError(t, err)

	newUser := domain.User{}
	json.Unmarshal(userJSON, &newUser)
	assert.Empty(t, newUser.Password)
}

func TestUser_Validation(t *testing.T) {
	var tests = []struct {
		user          domain.User
		expectedError error
	}{
		{domain.User{Password: "ewgflikhghewioghew"}, nil},
		{domain.User{}, domain.ErrUserMissingPassword},
	}

	for _, test := range tests {
		_, err := validate(test.user)

		if test.expectedError != nil {
			assert.EqualErrorf(t, err, test.expectedError.Error(), "expected user.Validate of %v to be %v, got %v instead", test.user, test.expectedError, err)
		} else {
			assert.NoErrorf(t, err, "expected user.Validate of %v to be %v, got %v instead", test.user, test.expectedError, err)
		}
	}
}
