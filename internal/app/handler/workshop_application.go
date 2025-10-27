package handler

import (
	"Iu5-web/internal/app/api_types"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetProductionInfo(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	appID, count, err := h.Repository.GetCartInfo(userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api_types.CartInfoResponse{
		ApplicationID: appID,
		ItemCount:     count,
	})
}

func (h *Handler) GetWorkshopApplications(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	isModerator, _ := GetUserRole(c)
	status := c.Query("status")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")
	var dateFrom, dateTo time.Time
	if dateFromStr != "" {
		dateFrom, err = time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from format, use YYYY-MM-DD"})
			return
		}
	}
	if dateToStr != "" {
		dateTo, err = time.Parse("2006-01-02", dateToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to format, use YYYY-MM-DD"})
			return
		}
	}
	apps, err := h.Repository.GetWorkshopApplications(status, dateFrom, dateTo, userID, isModerator)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	responses := make([]api_types.ApplicationResponse, len(apps))
	for i, app := range apps {
		count, err := h.Repository.GetApplicationItemsCount(app.ID)
		if err != nil {
			h.errorHandler(c, err)
			return
		}
		responses[i] = api_types.ConvertApplicationToResponse(app, count)
	}
	c.JSON(http.StatusOK, responses)
}

func (h *Handler) GetWorkshopApplicationByID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	isModerator, _ := GetUserRole(c)
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	app, items, err := h.Repository.GetWorkshopApplicationWithItems(uint(appID), userID, isModerator)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api_types.ConvertApplicationToDetailedResponse(app, items))
}

func (h *Handler) UpdateWorkshopApplication(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req api_types.ApplicationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	app, err := h.Repository.UpdateWorkshopApplication(uint(appID), userID, req)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	_, items, _ := h.Repository.GetWorkshopApplicationWithItems(app.ID, userID, false)
	c.JSON(http.StatusOK, api_types.ConvertApplicationToDetailedResponse(app, items))
}

func (h *Handler) FormWorkshopApplication(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	app, err := h.Repository.FormApplication(uint(appID), userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	_, items, _ := h.Repository.GetWorkshopApplicationWithItems(app.ID, userID, false)
	c.JSON(http.StatusOK, api_types.ConvertApplicationToDetailedResponse(app, items))
}

func (h *Handler) CompleteWorkshopApplication(c *gin.Context) {
	moderatorID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	app, err := h.Repository.CompleteApplication(uint(appID), moderatorID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	_, items, _ := h.Repository.GetWorkshopApplicationWithItems(app.ID, moderatorID, true)
	c.JSON(http.StatusOK, api_types.ConvertApplicationToDetailedResponse(app, items))
}

func (h *Handler) DeleteWorkshopApplication(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	appID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.Repository.DeleteWorkshopApplication(uint(appID), userID); err != nil {
		h.errorHandler(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
