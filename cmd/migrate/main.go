package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/draco777/gophermart/internal/storage"
)

func main() {
	var databaseURI string
	flag.StringVar(&databaseURI, "d", "", "Database URI")
	flag.Parse()

	// Переменные окружения имеют приоритет над флагами
	if dbURI := os.Getenv("DATABASE_URI"); dbURI != "" {
		databaseURI = dbURI
	}

	if databaseURI == "" {
		log.Fatal("Database URI is required. Use -d flag or DATABASE_URI environment variable")
	}

	log.Printf("Running migrations on database...")

	if err := storage.RunMigrations(databaseURI); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("Migrations completed successfully")
}
