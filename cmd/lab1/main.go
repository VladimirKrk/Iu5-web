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

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("ошибка инициализации конфига: %v", err)
	}

	postgresString := dsn.FromEnv()
	if postgresString == "" {
		logrus.Fatal("не удалось собрать DSN для подключения к БД")
	}

	repo, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("ошибка инициализации репозитория: %v", err)
	}

	hand := handler.NewHandler(repo)
	router := gin.Default()

	application := pkg.New(conf, router, hand)
	application.Run()

	logrus.Info("Application terminated")
}
