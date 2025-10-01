package main

import (
	"Iu5-web/internal/app/ds"
	"Iu5-web/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Загружаем переменные окружения из .env
	_ = godotenv.Load()

	// подключаемся к базе данных как в основном приложении
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("не удалось подключиться к базе данных")
	}

	// автомиграция gorm looks through the scheme and auto updates the database scheme.
	//creates tables , adds missing columns, indexes and so on
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Workshop{},
		&ds.ProductionOrder{},
		&ds.OrderWorkshop{},
	)
	if err != nil {
		panic("не удалось выполнить миграцию базы данных")
	}
}
