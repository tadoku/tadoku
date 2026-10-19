package main

import (
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	smtpinfra "github.com/tadoku/tadoku/services/tadoku-api/infra/smtp"
)

type emailConfig struct {
	Enabled       bool          `envconfig:"enabled" default:"false"`
	Address       string        `envconfig:"address"`
	Username      string        `envconfig:"username"`
	Password      string        `envconfig:"password"`
	Sender        string        `envconfig:"sender" default:"no-reply@tadoku.app"`
	SourceARN     string        `envconfig:"source_arn"`
	FromARN       string        `envconfig:"from_arn"`
	ReturnPathARN string        `envconfig:"return_path_arn"`
	Timeout       time.Duration `envconfig:"timeout" default:"10s"`
}

func (cfg emailConfig) sender() (email.Sender, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	return smtpinfra.New(smtpinfra.Config{
		Address: cfg.Address, Username: cfg.Username, Password: cfg.Password, Sender: cfg.Sender,
		SourceARN: cfg.SourceARN, FromARN: cfg.FromARN, ReturnPathARN: cfg.ReturnPathARN, Timeout: cfg.Timeout,
	})
}
