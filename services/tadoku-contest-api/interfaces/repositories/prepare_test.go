package repositories_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/infra"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/interfaces/repositories"
	"github.com/tadoku/api/test"

	txdb "github.com/DATA-DOG/go-txdb"
	"github.com/DavidHuie/gomigrate"
	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	cfg := loadConfig()

	// Must be called pgx so the sqlx mapper uses the correct notation
	txdb.Register("pgx", "postgres", cfg.DatabaseURL)

	db, err := infra.NewRDB(cfg.DatabaseURL, cfg.DatabaseMaxIdleConns, cfg.DatabaseMaxOpenConns)
	if err != nil {
		panic(fmt.Sprintf("could not connect to testing DB: %s", err))
	}

	migrator, _ := gomigrate.NewMigratorWithLogger(
		db.DB,
		gomigrate.Postgres{},
		"./../../migrations",
		log.New(ioutil.Discard, "", log.LstdFlags),
	)

	err = migrator.Migrate()
	if err != nil {
		panic(fmt.Sprintf("could not migrate testing DB: %s", err))
	}

	code := m.Run()
	defer os.Exit(code)
}

func loadConfig() *test.Config {
	c, err := test.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("could not load config: %s", err))
	}

	return c
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
