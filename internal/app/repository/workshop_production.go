package repository

import (
	"Iu5-web/internal/app/api_types"
	"Iu5-web/internal/app/ds"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type UpdatePredictionRequest struct {
	ApplicationID   uint   `json:"application_id"`
	WorkshopID      uint   `json:"workshop_id"`
	PredictedOutput string `json:"predicted_output"`
}

func (r *Repository) checkDraftAccess(applicationID, userID uint) error {
	var application ds.WorkshopApplication
	err := r.db.Where("id = ?", applicationID).First(&application).Error
	if err != nil {
		return fmt.Errorf("%w: application not found", ErrNotFound)
	}
	if application.Status != "draft" {
		return fmt.Errorf("%w: can only modify a draft application", ErrNotAllowed)
	}
	if application.CreatorID != userID {
		return fmt.Errorf("%w: you are not the creator of this draft", ErrNotAllowed)
	}
	return nil
}

func (r *Repository) AddWorkshopToApplication(workshopID, creatorID uint) (ds.WorkshopProduction, error) {
	if _, err := r.GetWorkshopByID(workshopID); err != nil {
		return ds.WorkshopProduction{}, err
	}
	draft, err := r.GetOrCreateDraftApplication(creatorID)
	if err != nil {
		return ds.WorkshopProduction{}, err
	}
	var existingLink ds.WorkshopProduction
	err = r.db.Where("application_id = ? AND workshop_id = ?", draft.ID, workshopID).First(&existingLink).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return ds.WorkshopProduction{}, fmt.Errorf("%w: this workshop is already in the application", ErrAlreadyExists)
		}
		return ds.WorkshopProduction{}, err
	}
	newLink := ds.WorkshopProduction{ApplicationID: draft.ID, WorkshopID: workshopID, FoundDefects: 0}
	if err := r.db.Create(&newLink).Error; err != nil {
		return ds.WorkshopProduction{}, err
	}
	if err := r.db.Preload("Workshop").First(&newLink, "application_id = ? AND workshop_id = ?", newLink.ApplicationID, newLink.WorkshopID).Error; err != nil {
		return ds.WorkshopProduction{}, err
	}
	return newLink, nil
}

func (r *Repository) DeleteWorkshopFromApplication(applicationID, workshopID, userID uint) error {
	if err := r.checkDraftAccess(applicationID, userID); err != nil {
		return err
	}
	result := r.db.Where("application_id = ? AND workshop_id = ?", applicationID, workshopID).Delete(&ds.WorkshopProduction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: workshop not found in this draft application", ErrNotFound)
	}
	return nil
}

func (r *Repository) UpdateProductionItem(applicationID, workshopID, userID uint, req api_types.ProductionItemUpdateRequest) (ds.WorkshopProduction, error) {
	if err := r.checkDraftAccess(applicationID, userID); err != nil {
		return ds.WorkshopProduction{}, err
	}
	if req.FoundDefects < 0 {
		return ds.WorkshopProduction{}, errors.New("found defects cannot be negative")
	}
	var item ds.WorkshopProduction
	err := r.db.Where("application_id = ? AND workshop_id = ?", applicationID, workshopID).First(&item).Error
	if err != nil {
		return ds.WorkshopProduction{}, fmt.Errorf("%w: workshop not found in this draft application", ErrNotFound)
	}
	item.FoundDefects = req.FoundDefects
	if err := r.db.Save(&item).Error; err != nil {
		return ds.WorkshopProduction{}, err
	}
	if err := r.db.Preload("Workshop").First(&item, "application_id = ? AND workshop_id = ?", item.ApplicationID, item.WorkshopID).Error; err != nil {
		return ds.WorkshopProduction{}, err
	}
	return item, nil
}

func (r *Repository) UpdatePrediction(appID, workshopID uint, result string) error {
	tx := r.db.Model(&ds.WorkshopProduction{}).
		Where("application_id = ? AND workshop_id = ?", appID, workshopID).
		Updates(map[string]interface{}{
			"predicted_output":   result,
			"calculation_status": "completed",
		})

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
