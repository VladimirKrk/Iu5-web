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
	logrus.Info("Starting server")

	api := a.Router.Group("/api")
	{
		// Роуты для Мастерских
		workshops := api.Group("/workshops")
		{
			workshops.GET("", a.Handler.GetWorkshops)
			workshops.GET("/:id", a.Handler.GetWorkshopByID)
			workshops.POST("", a.Handler.CreateWorkshop)
			workshops.PUT("/:id", a.Handler.UpdateWorkshop)
			workshops.DELETE("/:id", a.Handler.DeleteWorkshop)
			workshops.POST("/:id/image", a.Handler.UploadWorkshopImage)
		}

		// Роуты для Заказов
		orders := api.Group("/orders")
		{
			orders.GET("", a.Handler.GetApplications)
			orders.GET("/:id", a.Handler.GetApplicationByID)
			orders.PUT("/:id", a.Handler.UpdateApplication)
			orders.PUT("/:id/form", a.Handler.FormApplication)
			orders.PUT("/:id/complete", a.Handler.CompleteApplication)
			orders.DELETE("/:id", a.Handler.DeleteApplication)
		}

		// Роуты для Корзины
		cart := api.Group("/cart")
		{
			cart.GET("/icon", a.Handler.GetCartIcon)
			cart.POST("/workshops", a.Handler.AddWorkshopToCart)
			cart.PUT("/items", a.Handler.UpdateCartItem)
			cart.DELETE("/items", a.Handler.DeleteCartItem)
		}

		// Роуты для Пользователей
		api.POST("/register", a.Handler.RegisterUser)
		api.GET("/users/me", a.Handler.GetUserMe)
	}

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
}
