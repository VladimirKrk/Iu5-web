package main

import (
	"Iu5-web/internal/app/config"
	"Iu5-web/internal/app/dsn"
	"Iu5-web/internal/app/handler"
	"Iu5-web/internal/app/repository"
	"Iu5-web/internal/pkg"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Application start")

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		logrus.Fatal(err)
	}

	// --- НОВЫЙ БЛОК: ПОДКЛЮЧЕНИЕ К MINIO ---
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false // Мы работаем локально без HTTPS

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logrus.Fatalf("ошибка инициализации Minio: %v", err)
	}
	logrus.Info("Minio client initialized")
	// --- КОНЕЦ НОВОГО БЛОКА ---

	// Передаем Minio-клиент в обработчики
	hand := handler.NewHandler(repo, minioClient)

	router := gin.Default()
	application := pkg.New(conf, router, hand)
	application.Run()

	logrus.Info("Application terminated")
}
