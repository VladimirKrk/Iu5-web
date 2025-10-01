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

// Структуры для данных перед передачей в шаблон
type WorkshopForTmpl struct {
	ds.Workshop
	ImageURL      string
	ExtraImageURL string
}

type OrderForTmpl struct {
	ds.ProductionOrder
	Items []OrderWorkshopForTmpl
}

type OrderWorkshopForTmpl struct {
	ds.OrderWorkshop
	Workshop WorkshopForTmpl
}

// Обработчики (Handlers)

// GET / - Главная страница
func (h *Handler) GetWorkshopsPage(c *gin.Context) {
	searchQuery := c.Query("мастерская")
	var workshops []ds.Workshop
	var err error

	if searchQuery == "" {
		workshops, err = h.Repository.GetAllWorkshops()
	} else {
		workshops, err = h.Repository.GetWorkshopsByName(searchQuery)
	}
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка сервера")
		logrus.Error(err)
		return
	}

	workshopsForTmpl := make([]WorkshopForTmpl, len(workshops))
	for i, workshop := range workshops {
		workshopsForTmpl[i] = WorkshopForTmpl{
			Workshop: workshop,
			ImageURL: repository.MINIO_URL + workshop.ImageKey,
		}
	}

	draftOrder, _ := h.Repository.FindOrCreateDraftOrder(currentUserID)
	itemCount, _ := h.Repository.GetDraftOrderItemCount(currentUserID)

	c.HTML(http.StatusOK, "workshop_list.html", gin.H{
		"Workshops":   workshopsForTmpl,
		"Order":       draftOrder,
		"ItemCount":   itemCount,
		"SearchQuery": searchQuery,
	})
}

// GET /workshop/:id - Детальная страница
func (h *Handler) GetWorkshopDetailPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	workshop, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Страница не найдена")
		return
	}

	workshopForTmpl := WorkshopForTmpl{
		Workshop:      workshop,
		ImageURL:      repository.MINIO_URL + workshop.ImageKey,
		ExtraImageURL: repository.MINIO_URL + workshop.ExtraImageKey,
	}

	c.HTML(http.StatusOK, "workshop_detail.html", workshopForTmpl)
}

// GET /order/:id - Страница заказа
func (h *Handler) GetOrderPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	order, err := h.Repository.GetOrderByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Заказ не найден")
		return
	}
	if order.Status == "удалён" || order.CreatorID != currentUserID {
		c.String(http.StatusForbidden, "Доступ запрещен")
		return
	}

	orderForTmpl := OrderForTmpl{
		ProductionOrder: order,
		Items:           make([]OrderWorkshopForTmpl, len(order.Items)),
	}
	for i, item := range order.Items {
		orderForTmpl.Items[i] = OrderWorkshopForTmpl{
			OrderWorkshop: item,
			Workshop: WorkshopForTmpl{
				Workshop: item.Service, // GORM заполняет это поле Service
				ImageURL: repository.MINIO_URL + item.Service.ImageKey,
			},
		}
	}

	c.HTML(http.StatusOK, "order_detail.html", orderForTmpl)
}

// POST /add-to-order - Добавление мастерской в заказ
func (h *Handler) AddToOrder(c *gin.Context) {
	workshopID, _ := strconv.Atoi(c.PostForm("workshop_id"))

	draftOrder, err := h.Repository.FindOrCreateDraftOrder(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Ошибка при создании заказа")
		return
	}

	err = h.Repository.AddWorkshopToOrder(draftOrder.ID, uint(workshopID))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value") {
		c.String(http.StatusInternalServerError, "Ошибка при добавлении в заказ")
		return
	}

	c.Redirect(http.StatusFound, "/")
}

// POST /delete-order - Логическое удаление заказа
func (h *Handler) DeleteOrder(c *gin.Context) {
	orderID, _ := strconv.Atoi(c.PostForm("order_id"))
	err := h.Repository.DeleteOrderLogically(uint(orderID), currentUserID)
	if err != nil {
		c.String(http.StatusNotFound, "Не удалось удалить заказ")
		return
	}
	c.Redirect(http.StatusFound, "/")
}
