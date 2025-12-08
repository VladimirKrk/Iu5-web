package api_types

import (
	"Iu5-web/internal/app/ds"
	"time"
)

// --- Запросы ---

// ApplicationUpdateRequest описывает JSON для обновления полей черновика
type ApplicationUpdateRequest struct {
	ProductionName string `json:"production_name"`
}

// ProductionItemUpdateRequest описывает JSON для обновления данных о браке в позиции заявки
type ProductionItemUpdateRequest struct {
	FoundDefects int `json:"found_defects"`
}

// --- Ответы ---

// CartInfoResponse описывает JSON для иконки "корзины" (черновика)
type CartInfoResponse struct {
	ApplicationID uint  `json:"application_id"`
	ItemCount     int64 `json:"item_count"`
}

// ProductionItemResponse описывает одну мастерскую внутри ответа по заявке
type ProductionItemResponse struct {
	Workshop          WorkshopResponse `json:"workshop"`
	FoundDefects      int              `json:"found_defects"`
	PredictedOutput   *string          `json:"predicted_output,omitempty"`
	CalculationStatus string           `json:"calculation_status,omitempty"`
}

// ApplicationResponse описывает краткий JSON для одной заявки (для списков)
type ApplicationResponse struct {
	ID          uint          `json:"id"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	FormedAt    *time.Time    `json:"formed_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Creator     UserResponse  `json:"creator"`
	Moderator   *UserResponse `json:"moderator,omitempty"`
	ItemsCount  int64         `json:"items_count"` // Добавим кол-во позиций
}

// ApplicationDetailedResponse описывает полный JSON для одной заявки (для GET by ID)
type ApplicationDetailedResponse struct {
	ID             uint                     `json:"id"`
	Status         string                   `json:"status"`
	ProductionName *string                  `json:"production_name,omitempty"`
	CreatedAt      time.Time                `json:"created_at"`
	FormedAt       *time.Time               `json:"formed_at,omitempty"`
	CompletedAt    *time.Time               `json:"completed_at,omitempty"`
	Creator        UserResponse             `json:"creator"`
	Moderator      *UserResponse            `json:"moderator,omitempty"`
	Items          []ProductionItemResponse `json:"items"`
}

// --- Функции-конвертеры ---

func ConvertApplicationToResponse(app ds.WorkshopApplication, itemsCount int64) ApplicationResponse {
	var formedAt, completedAt *time.Time
	if app.FormedAt.Valid {
		formedAt = &app.FormedAt.Time
	}
	if app.CompletedAt.Valid {
		completedAt = &app.CompletedAt.Time
	}
	var moderator *UserResponse
	if app.ModeratorID.Valid {
		modResp := ConvertUserToResponse(app.Moderator)
		moderator = &modResp
	}

	return ApplicationResponse{
		ID:          app.ID,
		Status:      app.Status,
		CreatedAt:   app.CreatedAt,
		FormedAt:    formedAt,
		CompletedAt: completedAt,
		Creator:     ConvertUserToResponse(app.Creator),
		Moderator:   moderator,
		ItemsCount:  itemsCount,
	}
}

func ConvertApplicationToDetailedResponse(app ds.WorkshopApplication, items []ds.WorkshopProduction) ApplicationDetailedResponse {
	baseResponse := ConvertApplicationToResponse(app, int64(len(items)))

	var productionName *string
	if app.ProductionName.Valid {
		productionName = &app.ProductionName.String
	}

	itemResponses := make([]ProductionItemResponse, len(items))
	for i, item := range items {
		var predictedOutput *string
		if item.PredictedOutput.Valid {
			predictedOutput = &item.PredictedOutput.String
		}

		dbImageKey := item.Workshop.ImageKey
		dbExtraImageKey := item.Workshop.ExtraImageKey

		var imageKey, extraImageKey *string
		if dbImageKey.Valid {
			imageKey = &dbImageKey.String
		}
		if dbExtraImageKey.Valid {
			extraImageKey = &dbExtraImageKey.String
		}

		itemResponses[i] = ProductionItemResponse{
			Workshop: WorkshopResponse{
				ID:            item.Workshop.ID,
				Name:          item.Workshop.Name,
				Description:   item.Workshop.Description,
				Century:       item.Workshop.Century,
				ImageKey:      imageKey,
				ExtraImageKey: extraImageKey,
			},
			FoundDefects:      item.FoundDefects,
			PredictedOutput:   predictedOutput,
			CalculationStatus: item.CalculationStatus.String,
		}

	}

	return ApplicationDetailedResponse{
		ID:             app.ID,
		Status:         baseResponse.Status,
		ProductionName: productionName,
		CreatedAt:      baseResponse.CreatedAt,
		FormedAt:       baseResponse.FormedAt,
		CompletedAt:    baseResponse.CompletedAt,
		Creator:        baseResponse.Creator,
		Moderator:      baseResponse.Moderator,
		Items:          itemResponses,
	}
}
