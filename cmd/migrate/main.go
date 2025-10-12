package main

import (
	"Iu5-web/internal/app/ds"
	"Iu5-web/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("не удалось подключиться к базе данных")
	}

	// Выполняем автомиграцию для всех наших моделей
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Workshop{},
		&ds.WorkshopApplication{},
		&ds.WorkshopProduction{},
	)
	if err != nil {
		panic("не удалось выполнить миграцию базы данных")
	}
}
