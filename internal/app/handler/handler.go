package handler

import (
	"Iu5-web/internal/app/ds"
	"Iu5-web/internal/app/repository"
	"context"
	"errors"
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
	"gorm.io/gorm"
)

const currentUserID = 1

type Handler struct {
	Repository  *repository.Repository
	MinioClient *minio.Client
}

// DTO для отображения заявки в СПИСКЕ
type ApplicationListItemDTO struct {
	ID        uint      `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Creator   string    `json:"creator_login"`

	// НОВОЕ ВЫЧИСЛЯЕМОЕ ПОЛЕ
	CompletedItemsCount int `json:"completed_items_count"`
}

// DTO для мастерской внутри заказа
type WorkshopInAppDTO struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Century string `json:"century"`
}

type ProductionItemResponseDTO struct {
	ID              uint   `json:"id"`
	ApplicationID   uint   `json:"application_id"`
	WorkshopID      uint   `json:"workshop_id"`
	FoundDefects    int    `json:"found_defects"`
	PredictedOutput string `json:"predicted_output"`
}

// DTO для позиции в заказе
type ProductionItemDTO struct {
	Workshop        WorkshopInAppDTO `json:"workshop"`
	FoundDefects    int              `json:"found_defects"`
	PredictedOutput string           `json:"predicted_output"`
}

// DTO для детального просмотра заказа
type ApplicationDetailDTO struct {
	ID             uint                `json:"id"`
	Status         string              `json:"status"`
	ProductionName *string             `json:"production_name"`
	CreatorLogin   string              `json:"creator_login"`
	Items          []ProductionItemDTO `json:"items"`
}

// DTO для мастерской
type WorkshopListItemDTO struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Century  string `json:"century"`
	ImageKey string `json:"image_key"`
}

func NewHandler(r *repository.Repository, mc *minio.Client) *Handler {
	return &Handler{
		Repository:  r,
		MinioClient: mc,
	}
}

// Workshops

func (h *Handler) GetWorkshops(c *gin.Context) {
	workshops, _ := h.Repository.GetWorkshops(c.Query("name"))

	responseItems := make([]WorkshopListItemDTO, len(workshops))
	for i, ws := range workshops {
		responseItems[i] = WorkshopListItemDTO{
			ID:       ws.ID,
			Name:     ws.Name,
			Century:  ws.Century,
			ImageKey: ws.ImageKey,
		}
	}

	c.JSON(http.StatusOK, responseItems)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
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
	workshop, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "мастерская не найдена"})
		return
	}
	if err := h.Repository.DeleteWorkshop(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить мастерскую"})
		return
	}
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	if workshop.ImageKey != "" {
		_ = h.MinioClient.RemoveObject(context.Background(), bucketName, workshop.ImageKey, minio.RemoveObjectOptions{})
	}
	if workshop.ExtraImageKey != "" {
		_ = h.MinioClient.RemoveObject(context.Background(), bucketName, workshop.ExtraImageKey, minio.RemoveObjectOptions{})
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) AddWorkshopToProduction(c *gin.Context) {
	var addData struct {
		WorkshopID uint `json:"workshop_id"`
	}
	if err := c.ShouldBindJSON(&addData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	item, err := h.Repository.AddWorkshopToApplication(draftApp.ID, addData.WorkshopID)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{"error": "эта мастерская уже в заявке"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось добавить в заявку"})
		return
	}

	response := ProductionItemResponseDTO{
		ID:              item.ID,
		ApplicationID:   item.ApplicationID,
		WorkshopID:      item.WorkshopID,
		FoundDefects:    item.FoundDefects,
		PredictedOutput: item.PredictedOutput,
	}

	c.JSON(http.StatusCreated, response)
}
func (h *Handler) UploadWorkshopImage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	workshop, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "мастерская не найдена"})
		return
	}
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	file, err := c.FormFile("image")
	if err == nil {
		if workshop.ImageKey != "" {
			_ = h.MinioClient.RemoveObject(context.Background(), bucketName, workshop.ImageKey, minio.RemoveObjectOptions{})
		}
		imageKey := uuid.New().String() + filepath.Ext(file.Filename)
		fileContent, _ := file.Open()
		defer fileContent.Close()
		_, err = h.MinioClient.PutObject(context.Background(), bucketName, imageKey, fileContent, file.Size, minio.PutObjectOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить основное изображение"})
			return
		}
		workshop.ImageKey = imageKey
	}
	extraFile, err := c.FormFile("extra_image")
	if err == nil {
		if workshop.ExtraImageKey != "" {
			_ = h.MinioClient.RemoveObject(context.Background(), bucketName, workshop.ExtraImageKey, minio.RemoveObjectOptions{})
		}
		extraImageKey := uuid.New().String() + filepath.Ext(extraFile.Filename)
		extraFileContent, _ := extraFile.Open()
		defer extraFileContent.Close()
		_, err = h.MinioClient.PutObject(context.Background(), bucketName, extraImageKey, extraFileContent, extraFile.Size, minio.PutObjectOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить доп. изображение"})
			return
		}
		workshop.ExtraImageKey = extraImageKey
	}
	if err = h.Repository.UpdateWorkshop(&workshop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить мастерскую"})
		return
	}
	c.JSON(http.StatusOK, workshop)
}

// Workshop Applications

// handler.go

func (h *Handler) GetWorkshopApplications(c *gin.Context) {
	status, dateFromString, dateToString := c.Query("status"), c.Query("date_from"), c.Query("date_to")
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

	apps, err := h.Repository.GetWorkshopApplications(status, dateFrom, dateTo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить заявки"})
		return
	}

	responseItems := make([]ApplicationListItemDTO, len(apps))
	for i, app := range apps {
		// Для каждой заявки вызываем новый метод подсчета
		completedCount, _ := h.Repository.GetCompletedItemCount(app.ID)

		responseItems[i] = ApplicationListItemDTO{
			ID:                  app.ID,
			Status:              app.Status,
			CreatedAt:           app.CreatedAt,
			Creator:             app.Creator.Login,
			CompletedItemsCount: int(completedCount),
		}
	}

	c.JSON(http.StatusOK, responseItems)
}

func (h *Handler) GetWorkshopApplicationByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Repository.GetWorkshopApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}
	itemsDTO := make([]ProductionItemDTO, len(app.Items))
	for i, item := range app.Items {
		itemsDTO[i] = ProductionItemDTO{
			Workshop: WorkshopInAppDTO{
				ID:      item.Workshop.ID,
				Name:    item.Workshop.Name,
				Century: item.Workshop.Century,
			},
			FoundDefects:    item.FoundDefects,
			PredictedOutput: item.PredictedOutput,
		}
	}

	response := ApplicationDetailDTO{
		ID:             app.ID,
		Status:         app.Status,
		ProductionName: app.ProductionName,
		CreatorLogin:   app.Creator.Login,
		Items:          itemsDTO,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) UpdateWorkshopApplication(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var updateData struct {
		ProductionName string `json:"production_name"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	app, err := h.Repository.GetWorkshopApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}
	if app.CreatorID != currentUserID || app.Status != "черновик" {
		c.JSON(http.StatusForbidden, gin.H{"error": "нет прав для изменения"})
		return
	}
	app.ProductionName = &updateData.ProductionName

	if err := h.Repository.UpdateWorkshopApplication(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить заявку"})
		return
	}

	itemsDTO := make([]ProductionItemDTO, len(app.Items))
	for i, item := range app.Items {
		itemsDTO[i] = ProductionItemDTO{
			Workshop: WorkshopInAppDTO{
				ID:      item.Workshop.ID,
				Name:    item.Workshop.Name,
				Century: item.Workshop.Century,
			},
			FoundDefects:    item.FoundDefects,
			PredictedOutput: item.PredictedOutput,
		}
	}

	response := ApplicationDetailDTO{
		ID:             app.ID,
		Status:         app.Status,
		ProductionName: app.ProductionName,
		CreatorLogin:   app.Creator.Login,
		Items:          itemsDTO,
	}

	// Отдаем "чистый" DTO
	c.JSON(http.StatusOK, response)
}

func (h *Handler) FormWorkshopApplication(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Repository.GetWorkshopApplicationByID(uint(id))
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
	if err := h.Repository.UpdateWorkshopApplication(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сформировать заявку"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) CompleteWorkshopApplication(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Repository.GetWorkshopApplicationByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена"})
		return
	}
	if app.Status != "сформирован" {
		c.JSON(http.StatusForbidden, gin.H{"error": "можно завершить только сформированную заявку"})
		return
	}
	for i, item := range app.Items {
		app.Items[i].PredictedOutput = ds.CalculateProductionOutput(item.FoundDefects)
		if err := h.Repository.UpdateWorkshopProductionItem(&app.Items[i]); err != nil {
			logrus.Errorf("не удалось обновить позицию %d: %v", item.ID, err)
		}
	}
	app.Status = "завершён"
	now := time.Now()
	app.CompletedAt = &now
	if err := h.Repository.UpdateWorkshopApplication(&app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось завершить заявку"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) DeleteWorkshopApplication(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.Repository.DeleteApplicationLogically(uint(id), currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "заявка не найдена или нет прав"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetProductionInfo(c *gin.Context) {
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	itemCount, _ := h.Repository.GetDraftItemCount(currentUserID)
	c.JSON(http.StatusOK, gin.H{"application_id": draftApp.ID, "item_count": itemCount})
}

// Workshop Production

func (h *Handler) UpdateProductionItem(c *gin.Context) {
	var updateData struct {
		WorkshopID   uint `json:"workshop_id"`
		FoundDefects int  `json:"found_defects"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	item, err := h.Repository.GetWorkshopProductionItem(draftApp.ID, updateData.WorkshopID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "товар в черновике не найден"})
		return
	}
	item.FoundDefects = updateData.FoundDefects
	if err := h.Repository.UpdateWorkshopProductionItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить товар"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteProductionItem(c *gin.Context) {
	var deleteData struct {
		WorkshopID uint `json:"workshop_id"`
	}
	if err := c.ShouldBindJSON(&deleteData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	if err := h.Repository.DeleteWorkshopProductionItem(draftApp.ID, deleteData.WorkshopID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось удалить товар"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Users

func (h *Handler) RegisterUser(c *gin.Context) {
	var user ds.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные данные"})
		return
	}
	if user.Login == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "логин и пароль обязательны"})
		return
	}
	if err := h.Repository.CreateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать пользователя"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) AuthenticateUser(c *gin.Context) {
	var loginData struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	user, err := h.Repository.GetUserByLogin(loginData.Login)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный логин или пароль"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сервера"})
		return
	}
	if user.Password != loginData.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный логин или пароль"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "успешный вход", "token": "fake-jwt-token"})
}

func (h *Handler) GetUserMe(c *gin.Context) {
	user, err := h.Repository.GetUserByID(currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdateUserMe(c *gin.Context) {
	var updateData ds.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверные входные данные"})
		return
	}
	user, err := h.Repository.GetUserByID(currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
		return
	}
	user.Login = updateData.Login
	if updateData.Password != "" {
		user.Password = updateData.Password
	}
	if err := h.Repository.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось обновить профиль"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func (h *Handler) DeauthorizeUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "успешный выход"})
}
