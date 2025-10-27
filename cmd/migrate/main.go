// FILE: cmd/migrate/main.go
package main

import (
	"Iu5-web/internal/app/ds"
	"Iu5-web/internal/app/dsn"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("Loading .env file...")
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, continuing with environment variables")
	}

	dsnString := dsn.FromEnv()
	if dsnString == "" {
		log.Fatal("DSN string is not configured. Please check your .env file.")
	}

	log.Println("Connecting to database...")
	db, err := gorm.Open(postgres.Open(dsnString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Starting database migration...")
	// Указываем наши модели из проекта Iu5-web
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Workshop{},
		&ds.WorkshopApplication{},
		&ds.WorkshopProduction{}, // Важно: Убедитесь, что эта модель тоже здесь
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed successfully!")
}
