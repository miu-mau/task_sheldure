package main

import (
	"flag"
	"fmt"
	"log"

	"task_shelduler/internal/database"
)

func main() {
	var (
		dbPath = flag.String("db", "internal/migrations/data/task_scheduler.db", "path to SQLite database file")
	)
	flag.Parse()

	db, err := database.OpenDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	if err := database.RunMigrations(db, "internal/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("✓ Migrations applied successfully")
}
