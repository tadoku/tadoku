// Package app composes concrete feature services into application operations.
package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type Application struct {
	announcements *announcements.Service
	pages         *pages.Service
	posts         *posts.Service
	db            *pgxpool.Pool
	permissions   *permissions.Checker
}

func New(announcements *announcements.Service, pages *pages.Service, posts *posts.Service, db *pgxpool.Pool, permissions *permissions.Checker) *Application {
	return &Application{
		announcements: announcements,
		pages:         pages,
		posts:         posts,
		db:            db,
		permissions:   permissions,
	}
}
