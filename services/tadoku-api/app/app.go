// Package app composes concrete feature services into application operations.
package app

import "github.com/tadoku/tadoku/services/tadoku-api/features/content"

type Application struct {
	content *content.Service
}

func New(content *content.Service) *Application {
	return &Application{
		content: content,
	}
}
