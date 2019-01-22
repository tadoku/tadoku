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
