package health

import (
	"context"
	"database/sql"
)

// HealthChecker represents a dependency that can be health-checked.
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

// CheckResult holds the outcome of a single dependency check.
type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// ReadyzResponse is the JSON response for the readiness endpoint.
type ReadyzResponse struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks"`
}

type postgresChecker struct {
	name string
	db   *sql.DB
}

// NewPostgresChecker creates a HealthChecker for a Postgres connection.
// The name parameter identifies this check in the response (e.g. "postgres",
// "postgres-content") so services with multiple databases can distinguish them.
func NewPostgresChecker(name string, db *sql.DB) HealthChecker {
	return &postgresChecker{name: name, db: db}
}

func (c *postgresChecker) Name() string                    { return c.name }
func (c *postgresChecker) Check(ctx context.Context) error { return c.db.PingContext(ctx) }
