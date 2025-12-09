// FILE: internal/app/handler/workshop_application.go
package handler

import (
	"Iu5-web/internal/app/api_types"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetProductionInfo godoc
// @Summary      Get cart info
// @Description  Get current user's draft application ID and item count.
// @Tags         Applications
// @Produce      json
// @Success      200 {object} api_types.CartInfoResponse
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Security     BearerAuth
// @Router       /workshop_applications/info [get]
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

// GetWorkshopApplications godoc
// @Summary      Get list of applications
// @Description  Get applications. Regular users see their own, moderators see all. Excludes drafts and deleted.
// @Tags         Applications
// @Produce      json
// @Param        status    query     string false "Filter by status (e.g., 'formed', 'completed')"
// @Param        date_from query     string false "Start date for filtering (YYYY-MM-DD)"
// @Param        date_to   query     string false "End date for filtering (YYYY-MM-DD)"
// @Success      200 {array} api_types.ApplicationResponse
// @Failure      400 {object} api_types.ErrorResponse "Invalid date format"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Security     BearerAuth
// @Router       /workshop_applications [get]
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
		itemsCount, err := h.Repository.GetApplicationItemsCount(app.ID)
		if err != nil {
			h.errorHandler(c, err)
			return
		}

		// V-- ВЫЗЫВАЕМ НОВЫЙ МЕТОД ДЛЯ ПОДСЧЕТА ПРОГРЕССА --V
		calculatedCount, err := h.Repository.GetCalculatedItemsCount(app.ID)
		if err != nil {
			h.errorHandler(c, err)
			return
		}

		responses[i] = api_types.ConvertApplicationToResponse(app, itemsCount, calculatedCount)
	}
	c.JSON(http.StatusOK, responses)
}

// GetWorkshopApplicationByID godoc
// @Summary      Get an application by ID
// @Description  Get details of a single application. Users can only access their own, moderators can access any.
// @Tags         Applications
// @Produce      json
// @Param        id path int true "Application ID"
// @Success      200 {object} api_types.ApplicationDetailedResponse
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Failure      404 {object} api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshop_applications/{id} [get]
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

// UpdateWorkshopApplication godoc
// @Summary      Update a draft application
// @Description  Update fields of a draft application (e.g., production name). Only for the creator.
// @Tags         Applications
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Application ID"
// @Param        request body api_types.ApplicationUpdateRequest true "Fields to update"
// @Success      200 {object} api_types.ApplicationDetailedResponse
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden (not creator or not a draft)"
// @Failure      404 {object} api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshop_applications/{id} [put]
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

// FormWorkshopApplication godoc
// @Summary      Submit (form) a draft application
// @Description  Change the status of a draft application to 'formed'. Only for the creator.
// @Tags         Applications
// @Produce      json
// @Param        id path int true "Application ID"
// @Success      200 {object} api_types.ApplicationDetailedResponse
// @Failure      400 {object} api_types.ErrorResponse "Cannot form an empty application"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Failure      404 {object} api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshop_applications/{id}/form [post]
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

// CompleteWorkshopApplication godoc
// @Summary      Complete a formed application (Moderator only)
// @Description  Change status to 'completed', calculates production output. Requires moderator rights.
// @Tags         Applications
// @Produce      json
// @Param        id path int true "Application ID"
// @Success      200 {object} api_types.ApplicationDetailedResponse
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Failure      404 {object} api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshop_applications/{id}/complete [post]
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

// DeleteWorkshopApplication godoc
// @Summary      Delete a draft application
// @Description  Logically delete a draft application. Only for the creator.
// @Tags         Applications
// @Produce      json
// @Param        id path int true "Application ID"
// @Success      204 "No Content"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Failure      404 {object} api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshop_applications/{id} [delete]
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
