package handler

import (
	"Iu5-web/internal/app/api_types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AddWorkshopToProduction(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var req struct {
		WorkshopID uint `json:"workshop_id"`
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
