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

	for i := range app.Items {
		app.Items[i].ImageURL = repository.MINIO_URL + app.Items[i].ImageKey
	}

	ctx.HTML(http.StatusOK, "application_detail.html", app)
}
