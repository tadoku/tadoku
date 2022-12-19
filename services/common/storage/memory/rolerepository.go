package memory

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type RoleRepository struct {
	userRoles map[string]string // email -> role
}

type roleConfig struct {
	Users []struct {
		Email string `yaml:"email"`
		Role  string `yaml:"role"`
	} `yaml:"users"`
}

func NewRoleRepository(path string) *RoleRepository {
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

	return &RoleRepository{
		userRoles: userRoles,
	}
}

func (c *RoleRepository) GetRole(email string) string {
	if role, ok := c.userRoles[email]; ok {
		return role
	}

	return "user"
}
