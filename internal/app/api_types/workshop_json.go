package api_types

import "Iu5-web/internal/app/ds"

// WorkshopRequest описывает JSON для создания/обновления мастерской
type WorkshopRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Century     string `json:"century"`
}

// WorkshopResponse описывает JSON для ответа по одной мастерской
type WorkshopResponse struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Century       string  `json:"century"`
	ImageKey      *string `json:"image_key,omitempty"`
	ExtraImageKey *string `json:"extra_image_key,omitempty"`
}

// ConvertWorkshopToResponse преобразует ds.Workshop в WorkshopResponse
func ConvertWorkshopToResponse(ws ds.Workshop) WorkshopResponse {
	var imageKey, extraImageKey *string
	if ws.ImageKey.Valid {
		imageKey = &ws.ImageKey.String
	}
	if ws.ExtraImageKey.Valid {
		extraImageKey = &ws.ExtraImageKey.String
	}
	return WorkshopResponse{
		ID:            ws.ID,
		Name:          ws.Name,
		Description:   ws.Description,
		Century:       ws.Century,
		ImageKey:      imageKey,
		ExtraImageKey: extraImageKey,
	}
}

// ConvertWorkshopsToResponse преобразует срез ds.Workshop в срез WorkshopResponse
func ConvertWorkshopsToResponse(workshops []ds.Workshop) []WorkshopResponse {
	responses := make([]WorkshopResponse, len(workshops))
	for i, s := range workshops {
		responses[i] = ConvertWorkshopToResponse(s)
	}
	return responses
}
