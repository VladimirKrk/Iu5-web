package main

import (
	"Iu5-web/internal/app/config"
	"Iu5-web/internal/app/dsn"
	"Iu5-web/internal/app/handler"
	"Iu5-web/internal/app/repository"
	"Iu5-web/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Application start")

	// 1. Инициализируем конфиг
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("ошибка инициализации конфига: %v", err)
	}

	// 2. Получаем строку для подключения к БД
	postgresString := dsn.FromEnv()
	if postgresString == "" {
		logrus.Fatal("не удалось собрать DSN для подключения к БД")
	}

	// 3. Инициализируем репозиторий
	repo, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("ошибка инициализации репозитория: %v", err)
	}

	// 4. Инициализируем обработчики
	hand := handler.NewHandler(repo)

	// 5. Инициализируем роутер
	router := gin.Default()

	// 6. Создаем и запускаем приложение
	application := pkg.New(conf, router, hand)
	application.Run()

	logrus.Info("Application terminated")
}
