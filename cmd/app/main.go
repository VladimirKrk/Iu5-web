package main

import (
	"Iu5-web/internal/app/config"
	"Iu5-web/internal/app/dsn"
	"Iu5-web/internal/app/handler"
	"Iu5-web/internal/app/minioClient"
	"Iu5-web/internal/app/repository"
	"Iu5-web/internal/pkg"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.Info("Application start")

	if err := godotenv.Load(".env"); err != nil {
		logrus.WithError(err).Warn("Warning: .env file not found or could not be loaded. Continuing with environment variables from OS.")
	} else {
		logrus.Info(".env file loaded successfully")
	}

	conf, err := config.NewConfig()

	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Инициализация клиента MinIO
	mc, err := minioClient.New()
	if err != nil {
		logrus.Fatalf("error initializing minio client: %v", err)
	}
	logrus.Info("Minio client initialized")

	// Инициализация репозитория (передаем DSN и клиент MinIO)
	repo, err := repository.New(dsn.FromEnv(), mc)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}
	logrus.Info("Repository initialized successfully.")

	// Создание хендлера (передаем только репозиторий)
	hand := handler.NewHandler(repo)

	router := gin.Default()
	router.Use(cors.Default())

	// Создание и запуск приложения
	application := pkg.New(conf, router, hand)
	application.Run()

	logrus.Info("Application terminated")
}
