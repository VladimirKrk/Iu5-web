package dsn

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// FromEnv читает переменные окружения из .env файла и формирует DSN строку
func FromEnv() string {
	// Загружаем переменные из .env файла в окружение
	if err := godotenv.Load(); err != nil {
		logrus.Warn("не удалось загрузить .env файл, используются системные переменные окружения")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	// Формируем DSN строку
	// Пример: "host=localhost port=5432 user=user_workshop password=password_workshop dbname=workshop_db sslmode=disable"
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname)
}
