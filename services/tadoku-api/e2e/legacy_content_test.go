package e2e_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/content-api/domain"
	"github.com/tadoku/tadoku/services/content-api/http/rest"
	"github.com/tadoku/tadoku/services/content-api/http/rest/openapi"
	"github.com/tadoku/tadoku/services/content-api/storage/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type legacyContentAPI struct {
	db      *sql.DB
	handler http.Handler
}

func newLegacyContentAPI(ctx context.Context, dsn string) (*legacyContentAPI, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(legacyClockConnector{stdlib.GetConnector(*config)})
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}

	repository := postgres.NewAnnouncementRepository(db)
	server := rest.NewServer(
		nil, nil, nil, nil, nil, nil, nil, nil, // Page operations are not exercised.
		nil, nil, nil, nil, nil, nil, nil, nil, // Post operations are not exercised.
		nil, nil, nil, nil, nil, // Other announcement operations are not exercised.
		domain.NewAnnouncementListActive(repository),
	)
	router := echo.New()
	router.Logger.SetOutput(io.Discard)
	// Reuse the production route registration with the facade's mount prefix.
	// Authentication and infrastructure middleware are outside this contract test.
	openapi.RegisterHandlersWithBaseURL(router, server, "/content")

	return &legacyContentAPI{db: db, handler: router}, nil
}

// The legacy query reads PostgreSQL's now(), not the application clock. Bind
// that input to the scenario's frozen time without changing production code,
// shifting fixture timestamps, or normalizing the returned HTTP response.
// This deliberately supports only the active-announcements query.
type legacyClockConnector struct {
	driver.Connector
}

func (c legacyClockConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &legacyClockConn{conn.(*stdlib.Conn)}, nil
}

type legacyClockConn struct {
	*stdlib.Conn
}

func (c *legacyClockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if !strings.HasPrefix(query, "-- name: ListActiveAnnouncements :many\n") ||
		strings.Count(query, "now()") != 2 || len(args) != 1 {
		return nil, errors.New("legacy parity clock: unexpected query or arguments; review clock binding")
	}

	query = strings.ReplaceAll(query, "now()", "$2::timestamptz")
	args = []driver.NamedValue{args[0], {Ordinal: 2, Value: timex.Now()}}
	return c.Conn.QueryContext(ctx, query, args)
}

func TestLegacyClockFollowsScenarioTime(t *testing.T) {
	active := filepath.Join("testdata", APITestName("ListActiveAnnouncements", http.StatusOK, "without", "auth"))
	empty := filepath.Join("testdata", APITestName("ListActiveAnnouncements", http.StatusOK, "without", "announcements"))
	reset(t, filepath.Join(active, "setup.sql"))

	timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
		checkHTTPGolden(t, legacyContent.handler, active, http.StatusOK)
	})

	// Reuse the same connection and seeded row. It expires at 13:00:00;
	// a cached clock/query argument would incorrectly return it again.
	timex.TheWorld(time.Date(2026, 9, 12, 13, 0, 0, 0, time.UTC), func() {
		checkHTTPGolden(t, legacyContent.handler, empty, http.StatusOK)
	})
}

func TestLegacyClockRejectsUnexpectedQueries(t *testing.T) {
	for _, test := range []struct {
		name  string
		query string
		args  []driver.NamedValue
	}{
		{
			name:  "different operation",
			query: "select now(), now()",
			args:  []driver.NamedValue{{Ordinal: 1, Value: "main"}},
		},
		{
			name:  "changed clock",
			query: "-- name: ListActiveAnnouncements :many\nselect current_timestamp",
			args:  []driver.NamedValue{{Ordinal: 1, Value: "main"}},
		},
		{
			name:  "changed arguments",
			query: "-- name: ListActiveAnnouncements :many\nselect now(), now()",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			// No connection: unsupported shapes must fail before reaching PostgreSQL.
			_, err := (&legacyClockConn{}).QueryContext(t.Context(), test.query, test.args)
			if err == nil {
				t.Error("unexpected query accepted")
			}
		})
	}
}
