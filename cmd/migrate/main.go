package main

import (
	"log"

	"github.com/DavidHuie/gomigrate"

	"github.com/tadoku/api/app"
)

func main() {
	deps := app.NewServerDependencies()
	deps.AutoConfigure()

	db := deps.RDB().DB
	migrator, _ := gomigrate.NewMigrator(db, gomigrate.Postgres{}, "./migrations")

	err := migrator.Migrate()
	if err != nil {
		log.Printf("Error during migration: %v\n", err)
		log.Printf("Migration failed, trying to roll back...\n")

		err = migrator.Rollback()
		if err != nil {
			log.Fatalf("Migration is seriously broken: %v\n", err)
		}
	}
}
