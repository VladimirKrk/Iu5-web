// FILE: internal/app/handler/workshop_production.go
package handler

import (
	"Iu5-web/internal/app/api_types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddWorkshopToProduction godoc
// @Summary      Add a workshop to the draft application
// @Description  Adds a workshop to the current user's draft.
// @Tags         Production (Cart)
// @Accept       json
// @Produce      json
// @Param        request body object{workshop_id=int} true "Workshop ID to add"
// @Success      201 {object} api_types.ProductionItemResponse
// @Failure      400 {object} api_types.ErrorResponse "Invalid input"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      404 {object} api_types.ErrorResponse "Workshop not found"
// @Failure      409 {object} api_types.ErrorResponse "Workshop already in draft"
// @Security     BearerAuth
// @Router       /workshop_production/items [post]
func (h *Handler) AddWorkshopToProduction(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		WorkshopID uint `json:"workshop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	item, err := h.Repository.AddWorkshopToApplication(req.WorkshopID, userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	response := api_types.ProductionItemResponse{
		Workshop:     api_types.ConvertWorkshopToResponse(item.Workshop),
		FoundDefects: item.FoundDefects,
	}
	c.JSON(http.StatusCreated, response)
}

// UpdateProductionItem godoc
// @Summary      Update an item in a draft application
// @Description  Changes 'found_defects' for a workshop within the user's draft.
// @Tags         Production (Cart)
// @Accept       json
// @Produce      json
// @Param        app_id  path      int  true  "Application ID"
// @Param        ws_id   path      int  true  "Workshop ID"
// @Param        request body api_types.ProductionItemUpdateRequest true "Parameters to update"
// @Success      200 {object} api_types.ProductionItemResponse
// @Failure      400 {object} api_types.ErrorResponse "Invalid input"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden (not creator or not a draft)"
// @Failure      404 {object} api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshop_production/{app_id}/items/{ws_id} [put]
func (h *Handler) UpdateProductionItem(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	appID, _ := strconv.ParseUint(c.Param("app_id"), 10, 32)
	wsID, _ := strconv.ParseUint(c.Param("ws_id"), 10, 32)
	var req api_types.ProductionItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	item, err := h.Repository.UpdateProductionItem(uint(appID), uint(wsID), userID, req)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	response := api_types.ProductionItemResponse{
		Workshop:     api_types.ConvertWorkshopToResponse(item.Workshop),
		FoundDefects: item.FoundDefects,
	}
	c.JSON(http.StatusOK, response)
}

// DeleteProductionItem godoc
// @Summary      Delete an item from a draft application
// @Description  Removes a workshop from the current user's draft.
// @Tags         Production (Cart)
// @Produce      json
// @Param        app_id  path      int  true  "Application ID"
// @Param        ws_id   path      int  true  "Workshop ID"
// @Success      204 "No Content"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Failure      404 {object} api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshop_production/{app_id}/items/{ws_id} [delete]
func (h *Handler) DeleteProductionItem(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	appID, _ := strconv.ParseUint(c.Param("app_id"), 10, 32)
	wsID, _ := strconv.ParseUint(c.Param("ws_id"), 10, 32)
	if err := h.Repository.DeleteWorkshopFromApplication(uint(appID), uint(wsID), userID); err != nil {
		h.errorHandler(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RejectWorkshopApplication godoc
// @Summary      Reject a formed application (Moderator only)
// @Description  Changes the status of a formed application to 'rejected'. Requires moderator rights.
// @Tags         Applications
// @Produce      json
// @Param        id   path      int  true  "Application ID"
// @Success      200  {object}  api_types.ApplicationResponse
// @Failure      401  {object}  api_types.ErrorResponse "Unauthorized"
// @Failure      403  {object}  api_types.ErrorResponse "Forbidden (not moderator or not a 'formed' application)"
// @Failure      404  {object}  api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshop_applications/{id}/reject [post]
func (h *Handler) RejectWorkshopApplication(c *gin.Context) {
	appID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	app, err := h.Repository.RejectApplication(uint(appID))
	if err != nil {
		h.errorHandler(c, err)
		return
	}

	// Получаем количество элементов для корректного ответа
	count, _ := h.Repository.GetApplicationItemsCount(app.ID)
	c.JSON(http.StatusOK, api_types.ConvertApplicationToResponse(app, count))
}
