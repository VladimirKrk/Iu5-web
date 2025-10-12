package handler

import (
	"Iu5-web/internal/app/ds"
	"Iu5-web/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const currentUserID = 1

type Handler struct{ Repository *repository.Repository }

func NewHandler(r *repository.Repository) *Handler { return &Handler{Repository: r} }

type WorkshopForTmpl struct {
	ds.Workshop
	ImageURL      string
	ExtraImageURL string
}

type OrderForTmpl struct {
	ProductionOrder ds.WorkshopApplication // Поле с явным именем
	Items           []ProductionForTmpl
}

type ProductionForTmpl struct {
	ds.WorkshopProduction
	Workshop WorkshopForTmpl
}

func (h *Handler) GetWorkshopsPage(c *gin.Context) {
	searchQuery := c.Query("мастерская")
	workshops, _ := h.Repository.GetWorkshopsByName(searchQuery)
	workshopsForTmpl := make([]WorkshopForTmpl, len(workshops))
	for i, ws := range workshops {
		workshopsForTmpl[i] = WorkshopForTmpl{
			Workshop: ws,
			ImageURL: repository.MINIO_URL + ws.ImageKey,
		}
	}
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	itemCount, _ := h.Repository.GetDraftApplicationItemCount(currentUserID)
	c.HTML(http.StatusOK, "workshop_list.html", gin.H{
		"Workshops":   workshopsForTmpl,
		"Application": draftApp,
		"ItemCount":   itemCount,
		"SearchQuery": searchQuery,
	})
}

func (h *Handler) GetWorkshopDetailPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ws, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Страница не найдена")
		return
	}
	wsForTmpl := WorkshopForTmpl{
		Workshop:      ws,
		ImageURL:      repository.MINIO_URL + ws.ImageKey,
		ExtraImageURL: repository.MINIO_URL + ws.ExtraImageKey,
	}
	c.HTML(http.StatusOK, "workshop_detail.html", wsForTmpl)
}

func (h *Handler) GetOrderPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Repository.GetApplicationByID(uint(id))
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	if app.Status == "удалён" || app.CreatorID != currentUserID {
		c.Redirect(http.StatusFound, "/")
		return
	}
	appForTmpl := OrderForTmpl{
		ProductionOrder: app, // Используем явное имя
		Items:           make([]ProductionForTmpl, len(app.Items)),
	}
	for i, item := range app.Items {
		appForTmpl.Items[i] = ProductionForTmpl{
			WorkshopProduction: item,
			Workshop: WorkshopForTmpl{
				Workshop: item.Workshop,
				ImageURL: repository.MINIO_URL + item.Workshop.ImageKey,
			},
		}
	}
	c.HTML(http.StatusOK, "order_detail.html", appForTmpl)
}

func (h *Handler) AddToOrder(c *gin.Context) {
	workshopID, _ := strconv.Atoi(c.PostForm("workshop_id"))
	draftApp, _ := h.Repository.FindOrCreateDraftApplication(currentUserID)
	_ = h.Repository.AddWorkshopToApplication(draftApp.ID, uint(workshopID))
	c.Redirect(http.StatusFound, "/")
}

func (h *Handler) DeleteOrder(c *gin.Context) {
	appID, _ := strconv.Atoi(c.PostForm("application_id"))
	_ = h.Repository.DeleteApplicationLogically(uint(appID), currentUserID)
	c.Redirect(http.StatusFound, "/")
}

func (h *Handler) UpdateProductionName(c *gin.Context) {
	appID_str := c.PostForm("application_id")
	newName := c.PostForm("production_name")

	appID, _ := strconv.Atoi(appID_str)

	_ = h.Repository.UpdateProductionName(uint(appID), newName)

	c.Redirect(http.StatusFound, "/workshop_production/"+appID_str)
}
