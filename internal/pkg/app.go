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
		// Домен "Мастерские"
		workshops := api.Group("/workshops")
		{
			workshops.GET("", a.Handler.GetWorkshops)
			workshops.GET("/:id", a.Handler.GetWorkshopByID)
			workshops.POST("", a.Handler.CreateWorkshop)
			workshops.PUT("/:id", a.Handler.UpdateWorkshop)
			workshops.DELETE("/:id", a.Handler.DeleteWorkshop)
			workshops.POST("/:id/image", a.Handler.UploadWorkshopImage)
		}

		// Домен "Заявки"
		applications := api.Group("/applications")
		{
			applications.GET("", a.Handler.GetWorkshopApplications)
			applications.GET("/:id", a.Handler.GetWorkshopApplicationByID)
			applications.PUT("/:id", a.Handler.UpdateWorkshopApplication)
			applications.PUT("/:id/form", a.Handler.FormWorkshopApplication)
			applications.PUT("/:id/complete", a.Handler.CompleteWorkshopApplication)
			applications.DELETE("/:id", a.Handler.DeleteWorkshopApplication)
		}

		// Домен "Продукция в заказе"
		production := api.Group("/production")
		{
			production.GET("/info", a.Handler.GetProductionInfo)
			production.POST("/items", a.Handler.AddWorkshopToProduction)
			production.PUT("/items", a.Handler.UpdateProductionItem)
			production.DELETE("/items", a.Handler.DeleteProductionItem)
		}

		// Домен "Пользователи"
		users := api.Group("/users")
		{
			users.GET("/me", a.Handler.GetUserMe)
			users.PUT("/me", a.Handler.UpdateUserMe)
		}
		api.POST("/register", a.Handler.RegisterUser)
		api.POST("/login", a.Handler.AuthenticateUser)
		api.POST("/logout", a.Handler.DeauthorizeUser)
	}

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
}
