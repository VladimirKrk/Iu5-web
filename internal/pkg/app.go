package pkg

import (
	"Iu5-web/internal/app/config"
	"Iu5-web/internal/app/handler"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type App struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func New(c *config.Config, r *gin.Engine, h *handler.Handler) *App {
	return &App{Config: c, Router: r, Handler: h}
}

func (a *App) Run() {
	logrus.Info("Server starting")

	// Регистрируем статику
	a.Router.LoadHTMLGlob("templates/*")
	a.Router.Static("/resources", "./resources")

	// Регистрируем роуты
	a.Router.GET("/", a.Handler.GetWorkshopsPage)
	a.Router.GET("/workshop/:id", a.Handler.GetWorkshopDetailPage)
	a.Router.GET("/мастерская/:id", a.Handler.GetOrderPage)

	a.Router.POST("/add-to-order", a.Handler.AddToOrder)
	a.Router.POST("/delete-order", a.Handler.DeleteOrder)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
}
