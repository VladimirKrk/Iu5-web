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

	// ВАЖНО: Настраиваем раздачу статики из вашей папки 'resources'
	// URL будет /resources/styles/style.css
	r.Static("/resources", "./resources")

	// Настраиваем роуты для вашей лабы
	r.GET("/", handler.ServiceListHandler)
	r.GET("/service/:id", handler.ServiceDetailHandler)
	r.GET("/application/:id", handler.ApplicationDetailHandler)

	r.Run()
	log.Println("Server down")
}
