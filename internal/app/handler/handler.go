package handler

import (
	"lab1/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// Обработчик для главной страницы со списком услуг
func (h *Handler) ServiceListHandler(ctx *gin.Context) {
	searchQuery := ctx.Query("q") // Используем "q" как в методичке
	var services []repository.Service
	var err error

	if searchQuery == "" {
		services, err = h.Repository.GetServices()
	} else {
		services, err = h.Repository.GetServicesByName(searchQuery)
	}
	if err != nil {
		logrus.Error(err)
	}

	for i := range services {
		services[i].ImageURL = repository.MINIO_URL + services[i].ImageKey
	}

	app, _ := h.Repository.GetApplication()

	ctx.HTML(http.StatusOK, "service_list.html", gin.H{
		"Services":    services,
		"Application": app,
		"SearchQuery": searchQuery,
	})
}

// Обработчик для детальной страницы услуги
func (h *Handler) ServiceDetailHandler(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	service, err := h.Repository.GetServiceByID(id)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "404 page not found")
		return
	}

	// Формируем URL для обеих картинок
	service.ImageURL = repository.MINIO_URL + service.ImageKey
	service.ExtraImageURL = repository.MINIO_URL + service.ExtraImageKey // <-- Вот это изменение

	ctx.HTML(http.StatusOK, "service_detail.html", service)
}

// Обработчик для страницы заявки (корзины)
func (h *Handler) ApplicationDetailHandler(ctx *gin.Context) {
	app, _ := h.Repository.GetApplication()

	// Создаем новую структуру для "обогащенных" данных, которые мы передадим в шаблон
	type EnrichedApplicationItem struct {
		Name            string
		Description     string
		Century         string
		FoundDefects    int
		PredictedOutput string
		ImageURL        string
	}

	// Создаем срез для хранения этих обогащенных данных
	enrichedItems := make([]EnrichedApplicationItem, 0, len(app.Items))

	// Проходим циклом по каждому элементу в заявке
	for _, item := range app.Items {
		// Для каждого элемента находим соответствующую услугу по ServiceID
		service, err := h.Repository.GetServiceByID(item.ServiceID)
		if err != nil {
			logrus.Error(err)
			continue // Пропускаем, если услуга не найдена
		}

		// Создаем обогащенный объект, объединяя данные из услуги и из заявки
		enrichedItem := EnrichedApplicationItem{
			Name:            service.Name,
			Description:     service.ShortDescription, // Используем короткое описание
			Century:         service.Century,
			FoundDefects:    item.FoundDefects,
			PredictedOutput: item.PredictedOutput,
			ImageURL:        repository.MINIO_URL + service.ImageKey,
		}
		// Добавляем его в наш срез
		enrichedItems = append(enrichedItems, enrichedItem)
	}

	// Передаем в шаблон ID заявки и новый, обогащенный список элементов
	ctx.HTML(http.StatusOK, "application_detail.html", gin.H{
		"ID":    app.ID,
		"Items": enrichedItems,
	})
}
