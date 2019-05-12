package repositories_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tadoku/api/infra"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/test"

	txdb "github.com/DATA-DOG/go-txdb"
	"github.com/DavidHuie/gomigrate"
	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	cfg := loadConfig()

	// Must be called pgx so the sqlx mapper uses the correct notation
	txdb.Register("pgx", "postgres", cfg.DatabaseURL)

	code := m.Run()
	defer os.Exit(code)
}

func loadConfig() *test.Config {
	c, err := test.LoadConfig()
	if err != nil {
		panic("could not load config")
	}

	return c
}

func setupTestingSuite(t *testing.T) (rdb.SQLHandler, func() error) {
	db, cleanup := prepareDB(t)

	migrator, _ := gomigrate.NewMigratorWithLogger(
		db.DB,
		gomigrate.Postgres{},
		"./../../migrations",
		log.New(ioutil.Discard, "", log.LstdFlags),
	)

	err := migrator.Migrate()
	if err != nil {
		t.Fatalf("could not migrate testing DB: %s", err)
	}

	return infra.NewSQLHandler(db), cleanup
}

func prepareDB(t *testing.T) (db *sqlx.DB, cleanup func() error) {
	cName := fmt.Sprintf("connection_%d", time.Now().UnixNano())
	db, err := sqlx.Open("pgx", cName)

	if err != nil {
		t.Fatalf("open pgx connection: %s", err)
	}

	return db, func() error {
		fmt.Printf("Closing: %s\n", cName)
		return db.Close()
	}
}
