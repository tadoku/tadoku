package test

import (
	"github.com/creasty/configo"

	"github.com/tadoku/api/domain"
)

// Config contains testing configuration data
type Config struct {
	DatabaseURL          string `envconfig:"database_url" valid:"required"`
	DatabaseMaxIdleConns int    `envconfig:"database_max_idle_conns" valid:"required"`
	DatabaseMaxOpenConns int    `envconfig:"database_max_open_conns" valid:"required"`
}

// LoadConfig helper
func LoadConfig() (*Config, error) {
	c := &Config{}
	opts := configo.Option{
		Prefix: "testing",
	}
	err := configo.Load(c, opts)

	return c, domain.WrapError(err)
}
