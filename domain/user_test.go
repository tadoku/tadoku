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
	// Happy path
	{
		u := User{
			Password: "ilkjgewojgewjpoe",
		}
		ok, err := u.Validate()
		assert.True(t, ok, "a new user with correct data should validate correctly")
		assert.NoError(t, err, "a new user with correct data should validate correctly")
	}

	// Missing password for new user
	{
		u := User{}
		ok, err := u.Validate()
		assert.False(t, ok, "a new user without a password should not validate correctly")
		assert.EqualError(t, err, ErrUserMissingPassword.Error(), "a new user without a password should not validate correctly")
	}
}
