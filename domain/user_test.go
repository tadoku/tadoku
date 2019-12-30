package domain_test

import (
	"encoding/json"
	"errors"
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
		// Password checks
		{
			domain.User{DisplayName: "foobar", Password: "apassword"},
			nil,
		},
		{
			domain.User{DisplayName: "foobar", Password: ""},
			domain.ErrUserMissingPassword,
		},

		// DisplayName checks
		{
			domain.User{DisplayName: "foobar123", Password: "apassword"},
			nil,
		},
		{
			domain.User{DisplayName: "神様", Password: "apassword"},
			nil,
		},
		{
			domain.User{DisplayName: "a", Password: "apassword"},
			errors.New("display_name: a does not validate as runelength(2|18)"),
		},
		{
			domain.User{DisplayName: "abcdefghijklmnopqrstuvwxyz", Password: "apassword"},
			errors.New("display_name: abcdefghijklmnopqrstuvwxyz does not validate as runelength(2|18)"),
		},
		{
			domain.User{DisplayName: "Robert'); DROP TABLE students;--", Password: "apassword"},
			errors.New("display_name: Robert'); DROP TABLE students;-- does not validate as utfletternum"),
		},
		{
			domain.User{DisplayName: "", Password: "apassword"},
			errors.New("display_name: non zero value required"),
		},
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
