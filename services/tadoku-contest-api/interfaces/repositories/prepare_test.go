package repositories_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/tadoku-contest-api/domain"
	"github.com/tadoku/tadoku/services/tadoku-contest-api/infra"
	"github.com/tadoku/tadoku/services/tadoku-contest-api/interfaces/rdb"
	"github.com/tadoku/tadoku/services/tadoku-contest-api/interfaces/repositories"

	txdb "github.com/DATA-DOG/go-txdb"
	"github.com/DavidHuie/gomigrate"
	"github.com/cenkalti/backoff"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMain(m *testing.M) {
	database := "database"

	// Spin up a postgres container for testing
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:11.6-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB": database,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		// Panic and fail since there isn't much we can do if the container doesn't start
		panic(fmt.Sprintf("could not create postgresql container: %s", err))
	}
	defer postgresC.Terminate(ctx)

	// Create connection string for test container
	host, err := postgresC.Host(ctx)
	if err != nil {
		panic(fmt.Sprintf("could not fetch postgresql container host: %s", err))
	}
	p, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		panic(fmt.Sprintf("could not fetch postgresql container port: %s", err))
	}
	databaseURL := fmt.Sprintf("host=%s port=%s user=postgres dbname=%s sslmode=disable", host, p.Port(), database)

	// Must be called pgx so the sqlx mapper uses the correct notation
	txdb.Register("pgx", "postgres", databaseURL)

	db, err := infra.NewRDB(databaseURL, 10, 10)
	if err != nil {
		panic(fmt.Sprintf("could not connect to testing DB: %s", err))
	}

	// Wait until database is ready before we migrate
	var pingDb backoff.Operation = func() error {
		err := db.DB.Ping()
		if err != nil {
			log.Println("DB is not ready...backing off...")
			return err
		}
		log.Println("DB is ready!")
		return nil
	}

	err = backoff.Retry(pingDb, backoff.NewExponentialBackOff())
	if err != nil {
		log.Panic(err)
	}

	migrator, _ := gomigrate.NewMigrator(
		db.DB,
		gomigrate.Postgres{},
		"./migrations",
	)

	err = migrator.Migrate()
	if err != nil {
		panic(fmt.Sprintf("could not migrate testing DB: %s", err))
	}

	fmt.Println("DB has been migrated, running test...")

	code := m.Run()
	defer os.Exit(code)
}

func setupTestingSuite(t *testing.T) (rdb.SQLHandler, func() error) {
	t.Parallel()

	db, cleanup := prepareDB(t)
	return infra.NewSQLHandler(db), cleanup
}

func prepareDB(t *testing.T) (db *sqlx.DB, cleanup func() error) {
	cName := fmt.Sprintf("connection_%d", time.Now().UnixNano())
	db, err := sqlx.Open("pgx", cName)

	if err != nil {
		t.Fatalf("open pgx connection: %s", err)
	}

	return db, db.Close
}

func createTestUsers(t *testing.T, sqlHandler rdb.SQLHandler, count int) []*domain.User {
	users := make([]*domain.User, count)

	repo := repositories.NewUserRepository(sqlHandler)

	for i := 0; i < count; i++ {
		user := &domain.User{
			Email:       fmt.Sprintf("foo+%d@bar.com", i),
			DisplayName: fmt.Sprintf("FOO %d", i),
			Password:    "foobar",
			Role:        domain.RoleUser,
			Preferences: &domain.Preferences{},
		}
		err := repo.Store(user)
		assert.NoError(t, err)

		users[i] = user
	}

	return users
}
