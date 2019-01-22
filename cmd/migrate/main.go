package main

import (
	"fmt"
	"log"

	"github.com/DavidHuie/gomigrate"

	"github.com/tadoku/api/app"
)

func main() {
	deps := app.NewServerDependencies()
	err := deps.AutoConfigure()
	if err != nil {
		panic(fmt.Sprintf("Migration failed: %v\n", err))
	}

	db := deps.RDB().DB
	migrator, _ := gomigrate.NewMigrator(db, gomigrate.Postgres{}, "./migrations")

	err = migrator.Migrate()
	if err != nil {
		log.Printf("Error during migration: %v\n", err)
		log.Printf("Migration failed, trying to roll back...\n")

		err = migrator.Rollback()
		if err != nil {
			log.Fatalf("Migration is seriously broken: %v\n", err)
		}
	}
}
