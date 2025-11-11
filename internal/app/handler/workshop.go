package handler

import (
	"Iu5-web/internal/app/api_types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetWorkshops godoc
// @Summary      Get list of workshops
// @Description  Get all workshops with optional name filter. Publicly accessible.
// @Tags         Workshops
// @Produce      json
// @Param        name query string false "Filter by workshop name (case-insensitive)"
// @Success      200 {array} api_types.WorkshopResponse
// @Failure      500 {object} api_types.ErrorResponse "Internal server error"
// @Router       /workshops [get]
func (h *Handler) GetWorkshops(c *gin.Context) {
	nameFilter := c.Query("name")
	workshops, err := h.Repository.GetWorkshops(nameFilter)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api_types.ConvertWorkshopsToResponse(workshops))
}

// GetWorkshopByID godoc
// @Summary      Get a workshop by ID
// @Description  Get details of a single workshop by its ID. Publicly accessible.
// @Tags         Workshops
// @Produce      json
// @Param        id   path      int  true  "Workshop ID"
// @Success      200  {object}  api_types.WorkshopResponse
// @Failure      400  {object}  api_types.ErrorResponse "Invalid ID format"
// @Failure      404  {object}  api_types.ErrorResponse "Not Found"
// @Router       /workshops/{id} [get]
func (h *Handler) GetWorkshopByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	workshop, err := h.Repository.GetWorkshopByID(uint(id))
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api_types.ConvertWorkshopToResponse(workshop))
}

// CreateWorkshop godoc
// @Summary      Create a new workshop (Moderator only)
// @Description  Adds a new workshop to the database. Requires moderator rights.
// @Tags         Workshops
// @Accept       json
// @Produce      json
// @Param        workshop body api_types.WorkshopRequest true "Workshop object"
// @Success      201 {object} api_types.WorkshopResponse
// @Failure      400 {object} api_types.ErrorResponse "Invalid input"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Security     BearerAuth
// @Router       /workshops [post]
func (h *Handler) CreateWorkshop(c *gin.Context) {
	var req api_types.WorkshopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	workshop, err := h.Repository.CreateWorkshop(req)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusCreated, api_types.ConvertWorkshopToResponse(workshop))
}

// UpdateWorkshop godoc
// @Summary      Update a workshop (Moderator only)
// @Description  Updates an existing workshop's data. Requires moderator rights.
// @Tags         Workshops
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Workshop ID"
// @Param        workshop body api_types.WorkshopRequest true "Workshop object"
// @Success      200 {object} api_types.WorkshopResponse
// @Failure      400 {object} api_types.ErrorResponse "Invalid input or ID"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Security     BearerAuth
// @Router       /workshops/{id} [put]
func (h *Handler) UpdateWorkshop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	var req api_types.WorkshopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	workshop, err := h.Repository.UpdateWorkshop(uint(id), req)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api_types.ConvertWorkshopToResponse(workshop))
}

// DeleteWorkshop godoc
// @Summary      Delete a workshop (Moderator only)
// @Description  Deletes a workshop by its ID. Requires moderator rights.
// @Tags         Workshops
// @Produce      json
// @Param        id   path      int  true  "Workshop ID"
// @Success      204 "No Content"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Failure      404 {object} api_types.ErrorResponse "Not Found"
// @Security     BearerAuth
// @Router       /workshops/{id} [delete]
func (h *Handler) DeleteWorkshop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	if err := h.Repository.DeleteWorkshop(uint(id)); err != nil {
		h.errorHandler(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UploadWorkshopImage godoc
// @Summary      Upload an image for a workshop (Moderator only)
// @Description  Upload an image and/or extra image for a workshop. Requires moderator rights.
// @Tags         Workshops
// @Accept       multipart/form-data
// @Produce      json
// @Param        id          path      int   true  "Workshop ID"
// @Param        image       formData  file  false "Main image file"
// @Param        extra_image formData  file  false "Extra image file"
// @Success      200 {object} api_types.WorkshopResponse
// @Failure      400 {object} api_types.ErrorResponse "Invalid ID or file"
// @Failure      401 {object} api_types.ErrorResponse "Unauthorized"
// @Failure      403 {object} api_types.ErrorResponse "Forbidden"
// @Security     BearerAuth
// @Router       /workshops/{id}/image [post]
func (h *Handler) UploadWorkshopImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	image, _ := c.FormFile("image")
	extraImage, _ := c.FormFile("extra_image")
	if image == nil && extraImage == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one image ('image' or 'extra_image') must be provided"})
		return
	}
	workshop, err := h.Repository.UploadWorkshopImage(uint(id), image, extraImage)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api_types.ConvertWorkshopToResponse(workshop))
}
