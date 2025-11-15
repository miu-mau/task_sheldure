package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"task_shelduler/internal/database"
)

func main() {
	var (
		dbPath        = flag.String("db", "internal/migrations/data/task_scheduler.db", "path to SQLite database file")
		migrationsDir = flag.String("dir", "internal/migrations", "directory with migration files")
		command       = flag.String("command", "up", "migration command: up, down, status, create")
		name          = flag.String("name", "", "name for new migration (used with create command)")
	)
	flag.Parse()

	db, err := database.OpenDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	switch *command {
	case "up":
		if err := database.RunMigrations(db, *migrationsDir); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("✓ Migrations applied successfully")
	case "status":
		// Можно добавить статус миграций позже
		fmt.Println("Status command not implemented yet")
	case "create":
		if *name == "" {
			log.Fatal("Name is required for create command")
		}
		// Создание новой миграции через goose CLI
		fmt.Printf("To create a new migration, run: goose -dir %s create %s sql", *migrationsDir, *name)
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		os.Exit(1)
	}
}
