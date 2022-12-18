package middleware

import (
	"fmt"
	"io/ioutil"

	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type RoleChecker struct {
	userRoles map[string]string // email -> role
}

type roleConfig struct {
	Users []struct {
		Email string `yaml:"email"`
		Role  string `yaml:"role"`
	} `yaml:"users"`
}

func NewRoleChecker(path string) *RoleChecker {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("unable to load list of user roles"))
	}

	roles := roleConfig{}
	if err := yaml.Unmarshal(f, &roles); err != nil {
		panic(fmt.Errorf("roles file is malformed: %w", err))
	}

	userRoles := map[string]string{}

	for _, user := range roles.Users {
		userRoles[user.Email] = user.Role
	}

	return &RoleChecker{
		userRoles: userRoles,
	}
}

func (c *RoleChecker) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}
