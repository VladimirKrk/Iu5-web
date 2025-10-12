package handler

import (
	"Iu5-web/internal/app/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const currentUserID = 1 // Заглушка для текущего пользователя

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// GET /api/workshops
func (h *Handler) GetWorkshops(c *gin.Context) {
	workshops, err := h.Repository.GetWorkshops(c.Query("name"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, workshops)
}

// GET /api/workshops/:id
func (h *Handler) GetWorkshopByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный id"})
		return
	}

	workshop, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "мастерская не найдена"})
		return
	}

	c.JSON(http.StatusOK, workshop)
}

// GET /api/orders
func (h *Handler) GetApplications(c *gin.Context) {
	status := c.Query("status")
	dateFromString := c.Query("date_from")
	dateToString := c.Query("date_to")

	var dateFrom, dateTo time.Time
	var err error
	if dateFromString != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат date_from, используйте YYYY-MM-DD"})
			return
		}
	}
	if dateToString != "" {
		dateTo, err = time.Parse("2006-01-02", dateToString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат date_to, используйте YYYY-MM-DD"})
			return
		}
	}

	apps, err := h.Repository.GetApplications(status, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить заявки"})
		logrus.Error(err)
		return
	}

	c.JSON(http.StatusOK, apps)
}

// GET /api/cart/icon
func (h *Handler) GetCartIcon(c *gin.Context) {
	// Ищем или создаем черновик для текущего пользователя
	draftApp, err := h.Repository.FindOrCreateDraftApplication(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения корзины"})
		logrus.Error(err)
		return
	}

	// Считаем количество товаров в этом черновике
	itemCount, err := h.Repository.GetDraftApplicationItemCount(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка подсчета товаров"})
		logrus.Error(err)
		return
	}

	// Возвращаем результат в формате JSON
	c.JSON(http.StatusOK, gin.H{
		"application_id": draftApp.ID,
		"item_count":     itemCount,
	})
}

// DELETE /api/orders/:id
func (h *Handler) DeleteApplication(c *gin.Context) {
	// Получаем id из URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный id"})
		return
	}

	// Вызываем метод репозитория для логического удаления
	// Передаем ID заявки и ID текущего пользователя для проверки прав
	err = h.Repository.DeleteApplicationLogically(uint(id), currentUserID)
	if err != nil {
		// if метод вернул gorm.ErrRecordNotFound,then заявка не найдена или не принадлежит пользователю
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена или у вас нет прав на ее удаление"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить заявку"})
		}
		logrus.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) CreateWorkshop(c *gin.Context)      { /* TODO */ }
func (h *Handler) UpdateWorkshop(c *gin.Context)      { /* TODO */ }
func (h *Handler) DeleteWorkshop(c *gin.Context)      { /* TODO */ }
func (h *Handler) UploadWorkshopImage(c *gin.Context) { /* TODO */ }
func (h *Handler) GetApplicationByID(c *gin.Context)  { /* TODO */ }
func (h *Handler) UpdateApplication(c *gin.Context)   { /* TODO */ }
func (h *Handler) FormApplication(c *gin.Context)     { /* TODO */ }
func (h *Handler) CompleteApplication(c *gin.Context) { /* TODO */ }
func (h *Handler) AddWorkshopToCart(c *gin.Context)   { /* TODO */ }
func (h *Handler) UpdateCartItem(c *gin.Context)      { /* TODO */ }
func (h *Handler) DeleteCartItem(c *gin.Context)      { /* TODO */ }
func (h *Handler) RegisterUser(c *gin.Context)        { /* TODO */ }
func (h *Handler) GetUserMe(c *gin.Context)           { /* TODO */ }
