package pkg

import (
	"Iu5-web/internal/app/config"
	"Iu5-web/internal/app/handler"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// App описывает основное приложение с конфигурацией, маршрутизатором и обработчиком
type App struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

// New создает новый экземпляр приложения
func New(c *config.Config, r *gin.Engine, h *handler.Handler) *App {
	return &App{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

// Run запускает веб-сервер и регистрирует маршруты
// @title           Workshop Production API
// @version         1.0
// @description     This is a service for calculating production output in workshops.
// @host            localhost:8080
// @BasePath        /api
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func (a *App) Run() {
	logrus.Info("Server start up")

	// Всю логику регистрации роутов делегируем хендлеру
	a.Handler.RegisterRoutes(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)

	logrus.Infof("Starting server at %s", serverAddress)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatalf("Failed to run server: %v", err)
	}

	logrus.Info("Server down")
}
