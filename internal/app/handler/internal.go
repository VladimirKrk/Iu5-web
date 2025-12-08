package handler

import (
	"Iu5-web/internal/app/repository" // Импорт нужен для доступа к типу
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdatePredictionResult(c *gin.Context) {
	var req repository.UpdatePredictionRequest // Используем тип из репозитория
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.Repository.UpdatePrediction(req.ApplicationID, req.WorkshopID, req.PredictedOutput)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
