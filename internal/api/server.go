package api

import (
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Fatalf("ошибка инициализации репозитория: %s", err.Error())
	}
	handler := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.Static("/resources", "./resources")

	// Настраиваем роуты
	r.GET("/", handler.ServiceListHandler)
	r.GET("/service/:id", handler.ServiceDetailHandler)
	// ИЗМЕНЕНО: /application/:id стало /мастерская/:id
	r.GET("/мастерская/:id", handler.ApplicationDetailHandler)

	r.Run()
	log.Println("Server down")
}
