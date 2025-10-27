package repository

import (
	"Iu5-web/internal/app/api_types"
	"Iu5-web/internal/app/ds"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) GetOrCreateDraftApplication(creatorID uint) (ds.WorkshopApplication, error) {
	var application ds.WorkshopApplication
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").First(&application).Error
	if err == nil {
		return application, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newDraft := ds.WorkshopApplication{Status: "draft", CreatorID: creatorID, CreatedAt: time.Now()}
		if errCreate := r.db.Create(&newDraft).Error; errCreate != nil {
			return ds.WorkshopApplication{}, errCreate
		}
		return newDraft, nil
	}
	return ds.WorkshopApplication{}, err
}

func (r *Repository) GetCartInfo(userID uint) (uint, int64, error) {
	if userID == 0 {
		return 0, 0, nil
	}
	draft, err := r.GetOrCreateDraftApplication(userID)
	if err != nil {
		return 0, 0, err
	}
	count, err := r.GetApplicationItemsCount(draft.ID)
	return draft.ID, count, err
}

func (r *Repository) GetApplicationItemsCount(appID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.WorkshopProduction{}).Where("application_id = ?", appID).Count(&count).Error
	return count, err
}

func (r *Repository) GetWorkshopApplications(status string, dateFrom, dateTo time.Time, userID uint, isModerator bool) ([]ds.WorkshopApplication, error) {
	var applications []ds.WorkshopApplication
	query := r.db.Preload("Creator").Preload("Moderator").Where("status NOT IN (?, ?)", "draft", "deleted")
	if !isModerator {
		query = query.Where("creator_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if !dateFrom.IsZero() {
		query = query.Where("formed_at >= ?", dateFrom)
	}
	if !dateTo.IsZero() {
		query = query.Where("formed_at < ?", dateTo.AddDate(0, 0, 1))
	}
	err := query.Order("id DESC").Find(&applications).Error
	return applications, err
}

func (r *Repository) GetWorkshopApplicationWithItems(appID, userID uint, isModerator bool) (ds.WorkshopApplication, []ds.WorkshopProduction, error) {
	var application ds.WorkshopApplication
	err := r.db.Preload("Creator").Preload("Moderator").First(&application, appID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.WorkshopApplication{}, nil, fmt.Errorf("%w: application with id %d", ErrNotFound, appID)
		}
		return ds.WorkshopApplication{}, nil, err
	}
	if !isModerator && application.CreatorID != userID {
		return ds.WorkshopApplication{}, nil, fmt.Errorf("%w: you don't have access to this application", ErrNotAllowed)
	}
	var items []ds.WorkshopProduction
	err = r.db.Preload("Workshop").Where("application_id = ?", appID).Find(&items).Error
	return application, items, err
}

func (r *Repository) UpdateWorkshopApplication(appID, userID uint, req api_types.ApplicationUpdateRequest) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	err := r.db.Where("id = ? AND status = 'draft'", appID).First(&app).Error
	if err != nil {
		return ds.WorkshopApplication{}, fmt.Errorf("%w: draft application not found", ErrNotFound)
	}
	if app.CreatorID != userID {
		return ds.WorkshopApplication{}, fmt.Errorf("%w: you are not the creator of this draft", ErrNotAllowed)
	}
	app.ProductionName = sql.NullString{String: req.ProductionName, Valid: true}
	if err := r.db.Save(&app).Error; err != nil {
		return ds.WorkshopApplication{}, err
	}
	return app, nil
}

func (r *Repository) FormApplication(appID, userID uint) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	if err := r.db.First(&app, appID).Error; err != nil {
		return ds.WorkshopApplication{}, fmt.Errorf("%w: application not found", ErrNotFound)
	}
	if app.CreatorID != userID {
		return ds.WorkshopApplication{}, fmt.Errorf("%w: only the creator can form an application", ErrNotAllowed)
	}
	if app.Status != "draft" {
		return ds.WorkshopApplication{}, fmt.Errorf("%w: can only form a 'draft' application", ErrNotAllowed)
	}
	count, err := r.GetApplicationItemsCount(appID)
	if err != nil {
		return ds.WorkshopApplication{}, err
	}
	if count == 0 {
		return ds.WorkshopApplication{}, errors.New("cannot form an empty application")
	}
	app.Status = "formed"
	app.FormedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err := r.db.Save(&app).Error; err != nil {
		return ds.WorkshopApplication{}, err
	}
	return app, nil
}

func (r *Repository) CompleteApplication(appID, moderatorID uint) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	err := r.db.First(&app, appID).Error
	if err != nil {
		return ds.WorkshopApplication{}, fmt.Errorf("%w: application not found", ErrNotFound)
	}
	if app.Status != "formed" {
		return ds.WorkshopApplication{}, fmt.Errorf("%w: can only complete a 'formed' application", ErrNotAllowed)
	}
	var items []ds.WorkshopProduction
	if err := r.db.Where("application_id = ?", appID).Find(&items).Error; err != nil {
		return ds.WorkshopApplication{}, err
	}
	err = r.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			predictedOutput := ds.CalculateProductionOutput(item.FoundDefects)
			err := tx.Model(&ds.WorkshopProduction{}).
				Where("application_id = ? AND workshop_id = ?", item.ApplicationID, item.WorkshopID).
				Update("predicted_output", predictedOutput).Error
			if err != nil {
				return err
			}
		}
		app.Status = "completed"
		app.ModeratorID = sql.NullInt64{Int64: int64(moderatorID), Valid: true}
		app.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		if err := tx.Save(&app).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return ds.WorkshopApplication{}, fmt.Errorf("failed to complete application: %w", err)
	}
	return app, nil
}

func (r *Repository) DeleteWorkshopApplication(appID, userID uint) error {
	var app ds.WorkshopApplication
	if err := r.db.First(&app, appID).Error; err != nil {
		return fmt.Errorf("%w: application not found", ErrNotFound)
	}
	if app.CreatorID != userID || app.Status != "draft" {
		return fmt.Errorf("%w: only creator can delete their own draft", ErrNotAllowed)
	}
	app.Status = "deleted"
	if err := r.db.Save(&app).Error; err != nil {
		return err
	}
	return nil
}
