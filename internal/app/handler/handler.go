package handler

import (
	"Iu5-web/internal/app/ds"
	"Iu5-web/internal/app/repository"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

const currentUserID = 1

type Handler struct {
	Repository  *repository.Repository
	MinioClient *minio.Client
}

func NewHandler(r *repository.Repository, mc *minio.Client) *Handler {
	return &Handler{
		Repository:  r,
		MinioClient: mc, // <-- Инициализируем новое поле
	}
}

//Workshops

func (h *Handler) GetWorkshops(c *gin.Context) {
	workshops, err := h.Repository.GetWorkshops(c.Query("name"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, workshops)
}

func (h *Handler) GetWorkshopByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	workshop, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "мастерская не найдена"})
		return
	}
	c.JSON(http.StatusOK, workshop)
}

func (h *Handler) CreateWorkshop(c *gin.Context) {
	var workshop ds.Workshop
	if err := c.ShouldBindJSON(&workshop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные: " + err.Error()})
		return
	}
	if err := h.Repository.CreateWorkshop(&workshop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать мастерскую"})
		return
	}
	c.JSON(http.StatusCreated, workshop)
}

func (h *Handler) UpdateWorkshop(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	workshop, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "мастерская не найдена"})
		return
	}
	if err := c.ShouldBindJSON(&workshop); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	if err := h.Repository.UpdateWorkshop(&workshop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить мастерскую"})
		return
	}
	c.JSON(http.StatusOK, workshop)
}

func (h *Handler) DeleteWorkshop(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// Здесь также должна быть логика удаления картинки из Minio
	if err := h.Repository.DeleteWorkshop(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить мастерскую"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) UploadWorkshopImage(c *gin.Context) {
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

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не загружен: " + err.Error()})
		return
	}

	// Генерируем уникальное имя для файла
	fileExt := filepath.Ext(file.Filename)
	imageKey := uuid.New().String() + fileExt

	// Открываем содержимое файла
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось прочитать файл"})
		return
	}
	defer fileContent.Close()

	// Загружаем файл в Minio
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	_, err = h.MinioClient.PutObject(context.Background(), bucketName, imageKey, fileContent, file.Size, minio.PutObjectOptions{
		ContentType: "application/octet-stream", // или более конкретный тип, если известен
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить файл в хранилище"})
		logrus.Error(err)
		return
	}

	// TODO: Здесь должна быть логика удаления старого файла из Minio, если он был

	// Обновляем запись в БД, сохраняя новое имя файла
	workshop.ImageKey = imageKey
	if err = h.Repository.UpdateWorkshop(&workshop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить мастерскую"})
		return
	}

	c.JSON(http.StatusOK, workshop)
}

//Orders

func (h *Handler) GetApplications(c *gin.Context) {
	status := c.Query("status")
	dateFromString, dateToString := c.Query("date_from"), c.Query("date_to")
	var dateFrom, dateTo time.Time
	var err error
	if dateFromString != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат date_from"})
			return
		}
	}
	if dateToString != "" {
		dateTo, err = time.Parse("2006-01-02", dateToString)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат date_to"})
			return
		}
	}
	apps, err := h.Repository.GetApplications(status, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить заявки"})
		return
	}
	c.JSON(http.StatusOK, apps)
}

func (h *Handler) GetApplicationByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Repository.GetApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}
	if app.Status == "удалён" || app.CreatorID != currentUserID {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена или нет доступа"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) UpdateApplication(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
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
	app.ProductionName = updateData.ProductionName
	if err := h.Repository.UpdateApplication(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить заявку"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) FormApplication(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Repository.GetApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}
	if app.CreatorID != currentUserID || app.Status != "черновик" {
		c.JSON(http.StatusForbidden, gin.H{"error": "можно сформировать только свой черновик"})
		return
	}
	if len(app.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "нельзя сформировать пустую заявку"})
		return
	}
	app.Status = "сформирован"
	now := time.Now()
	app.FormedAt = &now
	if err := h.Repository.UpdateApplication(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сформировать заявку"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) CompleteApplication(c *gin.Context) {
	// В реальном приложении здесь была бы проверка, что текущий пользователь - модератор
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Repository.GetApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}
	if app.Status != "сформирован" {
		c.JSON(http.StatusForbidden, gin.H{"error": "можно завершить только сформированную заявку"})
		return
	}

	// Вызываем формулу расчета
	for i, item := range app.Items {
		app.Items[i].PredictedOutput = ds.CalculateProductionOutput(item.FoundDefects)
		if err := h.Repository.UpdateApplicationItem(&app.Items[i]); err != nil {
			logrus.Error("не удалось обновить позицию заказа: ", err)
		}
	}

	app.Status = "завершён"
	now := time.Now()
	app.CompletedAt = &now
	// moderatorID := 2 // ID модератора, в реальном приложении из токена
	// app.ModeratorID = &moderatorID
	if err := h.Repository.UpdateApplication(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось завершить заявку"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) DeleteApplication(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.Repository.DeleteApplicationLogically(uint(id), currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена или нет прав"})
		return
	}
	c.Status(http.StatusNoContent)
}

//Cart

func (h *Handler) GetCartIcon(c *gin.Context) {
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	itemCount, _ := h.Repository.GetDraftApplicationItemCount(currentUserID)
	c.JSON(http.StatusOK, gin.H{
		"application_id": draftApp.ID,
		"item_count":     itemCount,
	})
}

func (h *Handler) AddWorkshopToCart(c *gin.Context) {
	workshopID, _ := strconv.Atoi(c.PostForm("workshop_id"))
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	item, err := h.Repository.AddWorkshopToApplication(draftApp.ID, uint(workshopID))
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "эта мастерская уже в заявке"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось добавить в заявку"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) UpdateCartItem(c *gin.Context) {
	var updateData struct {
		WorkshopID      uint   `json:"workshop_id"`
		FoundDefects    int    `json:"found_defects"`
		PredictedOutput string `json:"predicted_output"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	item, err := h.Repository.GetApplicationItem(draftApp.ID, updateData.WorkshopID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "товар в корзине не найден"})
		return
	}
	item.FoundDefects = updateData.FoundDefects
	item.PredictedOutput = updateData.PredictedOutput
	if err := h.Repository.UpdateApplicationItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить товар"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteCartItem(c *gin.Context) {
	var deleteData struct {
		WorkshopID uint `json:"workshop_id"`
	}
	if err := c.ShouldBindJSON(&deleteData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	if err := h.Repository.DeleteApplicationItem(draftApp.ID, deleteData.WorkshopID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить товар"})
		return
	}
	c.Status(http.StatusNoContent)
}

// User

// POST /api/register
func (h *Handler) RegisterUser(c *gin.Context) {
	var user ds.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		// пользователь существует
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "пользователь с таким логином уже существует"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось зарегистрировать пользователя"})
		logrus.Error(err)
		return
	}

	// Не возвращаем пароль
	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

// GET /api/users/me
func (h *Handler) GetUserMe(c *gin.Context) {
	user, err := h.Repository.GetUserByID(currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}

	// Не возвращаем пароль
	user.Password = ""
	c.JSON(http.StatusOK, user)
}
