// Package app composes concrete feature services into application operations.
package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type Application struct {
	content     *content.Service
	db          *pgxpool.Pool
	permissions *permissions.Checker
}

func New(content *content.Service, db *pgxpool.Pool, permissions *permissions.Checker) *Application {
	return &Application{
		content:     content,
		db:          db,
		permissions: permissions,
	}
}
