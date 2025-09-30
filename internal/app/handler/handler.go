package handler

import (
	"Iu5-web/internal/app/ds"
	"Iu5-web/internal/app/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const currentUserID = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// ==========================================================
// Структуры для "обогащения" данных перед передачей в шаблон
// ==========================================================
type ServiceForTmpl struct {
	ds.Service
	ImageURL      string
	ExtraImageURL string
}

type ApplicationForTmpl struct {
	ds.Application
	Items []ApplicationServiceForTmpl
}

type ApplicationServiceForTmpl struct {
	ds.ApplicationService
	Service ServiceForTmpl
}

// ==========================================================
// Обработчики (Handlers)
// ==========================================================

// GET / - Главная страница
func (h *Handler) GetServicesPage(c *gin.Context) {
	searchQuery := c.Query("мастерская")
	services, err := h.Repository.GetServicesByName(searchQuery) // Метод работает и для пустой строки
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка сервера")
		logrus.Error(err)
		return
	}

	// Обогащаем данные для шаблона
	servicesForTmpl := make([]ServiceForTmpl, len(services))
	for i, service := range services {
		servicesForTmpl[i] = ServiceForTmpl{
			Service:  service,
			ImageURL: repository.MINIO_URL + service.ImageKey,
		}
	}

	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	itemCount, _ := h.Repository.GetDraftApplicationItemCount(currentUserID)

	c.HTML(http.StatusOK, "service_list.html", gin.H{
		"Services":    servicesForTmpl,
		"Application": draftApp,
		"ItemCount":   itemCount,
		"SearchQuery": searchQuery,
	})
}

// GET /service/:id - Детальная страница
func (h *Handler) GetServiceDetailPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	service, err := h.Repository.GetServiceByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Страница не найдена")
		return
	}

	// Обогащаем данные для шаблона
	serviceForTmpl := ServiceForTmpl{
		Service:       service,
		ImageURL:      repository.MINIO_URL + service.ImageKey,
		ExtraImageURL: repository.MINIO_URL + service.ExtraImageKey,
	}

	c.HTML(http.StatusOK, "service_detail.html", serviceForTmpl)
}

// GET /мастерская/:id - Страница заявки
func (h *Handler) GetApplicationPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Repository.GetApplicationByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Заявка не найдена")
		return
	}
	if app.Status == "удалён" || app.CreatorID != currentUserID {
		c.String(http.StatusForbidden, "Доступ запрещен")
		return
	}

	// Обогащаем данные для шаблона
	appForTmpl := ApplicationForTmpl{
		Application: app,
		Items:       make([]ApplicationServiceForTmpl, len(app.Items)),
	}
	for i, item := range app.Items {
		appForTmpl.Items[i] = ApplicationServiceForTmpl{
			ApplicationService: item,
			Service: ServiceForTmpl{
				Service:  item.Service,
				ImageURL: repository.MINIO_URL + item.Service.ImageKey,
			},
		}
	}

	c.HTML(http.StatusOK, "application_detail.html", appForTmpl)
}

// POST /add-to-cart - Добавление услуги в заявку
func (h *Handler) AddToCart(c *gin.Context) {
	serviceID, _ := strconv.Atoi(c.PostForm("service_id"))

	draftApp, err := h.Repository.FindOrCreateDraftApplication(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка при создании заявки")
		logrus.Error(err)
		return
	}

	err = h.Repository.AddServiceToApplication(draftApp.ID, uint(serviceID))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value") {
		c.String(http.StatusInternalServerError, "Ошибка при добавлении услуги")
		logrus.Error(err)
		return
	}

	c.Redirect(http.StatusFound, "/")
}

// POST /delete-application - Логическое удаление заявки
func (h *Handler) DeleteApplication(c *gin.Context) {
	appID, _ := strconv.Atoi(c.PostForm("application_id"))
	err := h.Repository.DeleteApplicationLogically(uint(appID), currentUserID)
	if err != nil {
		c.String(http.StatusNotFound, "Не удалось удалить заявку")
		logrus.Error(err)
		return
	}
	c.Redirect(http.StatusFound, "/")
}
