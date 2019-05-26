package domain

import (
	"database/sql"
)

// ErrNotFound for when an entity could not be found
var ErrNotFound = sql.ErrNoRows
