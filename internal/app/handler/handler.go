package handler

import (
	"Iu5-web/internal/app/repository"
	"errors"
	"net/http"

	_ "Iu5-web/docs" // Import generated docs for Swagger

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterRoutes регистрирует все маршруты API
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		// --- Публичные роуты ---
		api.POST("/register", h.RegisterUser)
		api.POST("/login", h.AuthenticateUser)
		api.GET("/workshops", h.GetWorkshops)
		api.GET("/workshops/:id", h.GetWorkshopByID)

		// --- Защищенные роуты (требуют аутентификации) ---
		protected := api.Group("/")
		protected.Use(h.AuthMiddleware(false)) // false = не требует прав модератора
		{
			// Пользователь
			protected.POST("/logout", h.DeauthorizeUser)
			protected.GET("/users/me", h.GetUserMe)
			protected.PUT("/users/me", h.UpdateUserMe)

			// Работа с "корзиной" (черновиком заявки)
			protected.POST("/workshop_production/items", h.AddWorkshopToProduction)               // Добавить мастерскую
			protected.PUT("/workshop_production/:app_id/items/:ws_id", h.UpdateProductionItem)    // Обновить кол-во брака
			protected.DELETE("/workshop_production/:app_id/items/:ws_id", h.DeleteProductionItem) // Удалить мастерскую

			// Заявки
			protected.GET("/workshop_applications/info", h.GetProductionInfo) // "Корзина"
			protected.GET("/workshop_applications", h.GetWorkshopApplications)
			protected.GET("/workshop_applications/:id", h.GetWorkshopApplicationByID)
			protected.PUT("/workshop_applications/:id", h.UpdateWorkshopApplication)
			protected.POST("/workshop_applications/:id/form", h.FormWorkshopApplication)
			protected.DELETE("/workshop_applications/:id", h.DeleteWorkshopApplication)
		}

		// --- Роуты только для модераторов ---
		moderator := api.Group("/")
		moderator.Use(h.AuthMiddleware(true)) // true = требует прав модератора
		{
			// CRUD для мастерских
			moderator.POST("/workshops", h.CreateWorkshop)
			moderator.PUT("/workshops/:id", h.UpdateWorkshop)
			moderator.DELETE("/workshops/:id", h.DeleteWorkshop)
			moderator.POST("/workshops/:id/image", h.UploadWorkshopImage)

			// Завершение заявки
			moderator.POST("/workshop_applications/:id/complete", h.CompleteWorkshopApplication)
		}
		internal := api.Group("/internal")
		// Мы создадим это middleware в следующем шаге
		internal.Use(h.InternalAuthMiddleware())
		{
			internal.POST("/update-prediction", h.UpdatePredictionResult)
		}
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})
}

// errorHandler централизованно обрабатывает ошибки из репозитория
func (h *Handler) errorHandler(c *gin.Context, err error) {
	logrus.Error(err.Error())

	switch {
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrNotAllowed):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		// Для всех остальных ошибок возвращаем 500
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal server error occurred"})
	}
}
