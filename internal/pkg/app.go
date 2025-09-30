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

	a.Router.LoadHTMLGlob("templates/*")
	a.Router.Static("/resources", "./resources")

	// Используем правильные имена методов
	a.Router.GET("/", a.Handler.GetServicesPage)
	a.Router.GET("/service/:id", a.Handler.GetServiceDetailPage)
	a.Router.GET("/мастерская/:id", a.Handler.GetApplicationPage)

	a.Router.POST("/add-to-cart", a.Handler.AddToCart)
	a.Router.POST("/delete-application", a.Handler.DeleteApplication)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
}
