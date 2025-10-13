package handler

import (
	"Iu5-web/internal/app/ds"
	"Iu5-web/internal/app/repository"
	"net/http"
	"strconv"
	"strings"
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

// POST /api/workshops то есть делаем новюу мастерскую
func (h *Handler) CreateWorkshop(c *gin.Context) {
	var workshop ds.Workshop
	// Парсим JSON из тела запроса в нашу структуру
	if err := c.ShouldBindJSON(&workshop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные: " + err.Error()})
		return
	}

	// Вызываем метод для сохранения в БД
	if err := h.Repository.CreateWorkshop(&workshop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать мастерскую"})
		logrus.Error(err)
		return
	}

	// Возвращаем созданный объект и статус 201 Created
	c.JSON(http.StatusCreated, workshop)
}

// понадобится Minio пока стоит заглушка, просто обновляет имя файла в БД.
// POST /api/workshops/:id/image
func (h *Handler) UploadWorkshopImage(c *gin.Context) {
	// Получаем ID из URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный id"})
		return
	}

	// Находим мастерскую
	workshop, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "мастерская не найдена"})
		return
	}

	// Получаем файл из формы
	_, err = c.FormFile("image") // 'image' - это имя поля в multipart/form-data
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не загружен"})
		return
	}

	imageKey := "tulskiy_zavod.png"

	// Обновляем запись в БД
	workshop.ImageKey = imageKey
	if err = h.Repository.UpdateWorkshop(&workshop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить мастерскую"})
		logrus.Error(err)
		return
	}

	c.JSON(http.StatusOK, workshop)
}

// POST /api/cart/workshops
func (h *Handler) AddWorkshopToCart(c *gin.Context) {
	// Ожидаем, что ID мастерской будет передан в теле запроса в виде form-data
	workshopIDStr := c.PostForm("workshop_id")
	workshopID, err := strconv.Atoi(workshopIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный workshop_id"})
		return
	}

	// аходим или создаем заявку-черновик для текущего пользователя
	draftApp, err := h.Repository.FindOrCreateDraftApplication(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось найти или создать заявку"})
		logrus.Error(err)
		return
	}

	// Добавляем мастерскую в эту заявку
	item, err := h.Repository.AddWorkshopToApplication(draftApp.ID, uint(workshopID))
	if err != nil {
		// Обрабатываем ошибку, если такая услуга уже есть в заявке
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "эта мастерская уже добавлена в заявку"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось добавить мастерскую в заявку"})
		logrus.Error(err)
		return
	}

	// Возвращаем успешный ответ с информацией о созданной связи
	c.JSON(http.StatusCreated, item)
}

// GET /api/orders/:id
func (h *Handler) GetApplicationByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный id"})
		return
	}

	app, err := h.Repository.GetApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		logrus.Error(err)
		return
	}

	// Проверяем, что заявка не удалена и принадлежит текущему пользователю
	if app.Status == "удалён" || app.CreatorID != currentUserID {
		// Показываем ошибку 404
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}

	c.JSON(http.StatusOK, app)
}

// PUT /api/cart/items
func (h *Handler) UpdateCartItem(c *gin.Context) {
	// Ожидаем JSON с данными для обновления
	var updateData struct {
		WorkshopID      uint   `json:"workshop_id"`
		FoundDefects    int    `json:"found_defects"`
		PredictedOutput string `json:"predicted_output"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}

	// Находим черновик текущего пользователя
	draftApp, err := h.Repository.FindOrCreateDraftApplication(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось найти заявку"})
		return
	}

	// Находим конкретную позицию в заказе по ID заявки и ID мастерской
	item, err := h.Repository.GetApplicationItem(draftApp.ID, updateData.WorkshopID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "товар в корзине не найден"})
		return
	}

	// Обновляем поля
	item.FoundDefects = updateData.FoundDefects
	item.PredictedOutput = updateData.PredictedOutput

	// Сохраняем изменения в БД
	if err := h.Repository.UpdateApplicationItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить товар"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный id"})
		return
	}

	// Структура для парсинга JSON
	var updateData struct {
		ProductionName string `json:"production_name"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}

	app, err := h.Repository.GetApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}

	if app.CreatorID != currentUserID || app.Status != "черновик" {
		c.JSON(http.StatusForbidden, gin.H{"error": "нет прав для изменения этой заявки"})
		return
	}

	// Обновляем новое поле
	app.ProductionName = updateData.ProductionName

	if err := h.Repository.UpdateApplication(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить заявку"})
		return
	}

	c.JSON(http.StatusOK, app)
}

// PUT /api/orders/:id/form
func (h *Handler) FormApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный id"})
		return
	}

	app, err := h.Repository.GetApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}

	// только создатель может формировать свой черновик
	if app.CreatorID != currentUserID || app.Status != "черновик" {
		c.JSON(http.StatusForbidden, gin.H{"error": "можно сформировать только свою заявку в статусе 'черновик'"})
		return
	}

	if len(app.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нельзя сформировать пустую заявку"})
		return
	}

	// Меняем статус и ставим дату формирования
	app.Status = "сформирован"
	now := time.Now()
	app.FormedAt = &now // &now создает указатель на текущее время

	if err := h.Repository.UpdateApplication(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сформировать заявку"})
		return
	}

	c.JSON(http.StatusOK, app)
}

func (h *Handler) UpdateWorkshop(c *gin.Context)      { /* TODO */ }
func (h *Handler) DeleteWorkshop(c *gin.Context)      { /* TODO */ }
func (h *Handler) CompleteApplication(c *gin.Context) { /* TODO */ }
func (h *Handler) DeleteCartItem(c *gin.Context)      { /* TODO */ }
func (h *Handler) RegisterUser(c *gin.Context)        { /* TODO */ }
func (h *Handler) GetUserMe(c *gin.Context)           { /* TODO */ }
