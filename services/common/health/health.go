package health

import (
	"context"
	"database/sql"
)

type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}

type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type ReadyzResponse struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks"`
}

type postgresChecker struct {
	name string
	db   *sql.DB
}

func NewPostgresChecker(name string, db *sql.DB) HealthChecker {
	return &postgresChecker{name: name, db: db}
}

func (c *postgresChecker) Name() string                    { return c.name }
func (c *postgresChecker) Check(ctx context.Context) error { return c.db.PingContext(ctx) }
