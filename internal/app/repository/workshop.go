package repository

import (
	"Iu5-web/internal/app/api_types"
	"Iu5-web/internal/app/ds"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"

	"gorm.io/gorm"
)

func (r *Repository) GetWorkshops(nameFilter string) ([]ds.Workshop, error) {
	var workshops []ds.Workshop
	query := r.db.Order("name ASC")
	if nameFilter != "" {
		query = query.Where("name ILIKE ?", "%"+nameFilter+"%")
	}
	if err := query.Find(&workshops).Error; err != nil {
		return nil, err
	}
	return workshops, nil
}

func (r *Repository) GetWorkshopByID(id uint) (ds.Workshop, error) {
	var workshop ds.Workshop
	if err := r.db.First(&workshop, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Workshop{}, fmt.Errorf("%w: workshop with id %d", ErrNotFound, id)
		}
		return ds.Workshop{}, err
	}
	return workshop, nil
}

func (r *Repository) CreateWorkshop(req api_types.WorkshopRequest) (ds.Workshop, error) {
	newWorkshop := ds.Workshop{
		Name:        req.Name,
		Description: req.Description,
		Century:     req.Century,
	}
	if err := r.db.Create(&newWorkshop).Error; err != nil {
		return ds.Workshop{}, err
	}
	return newWorkshop, nil
}

func (r *Repository) UpdateWorkshop(id uint, req api_types.WorkshopRequest) (ds.Workshop, error) {
	workshop, err := r.GetWorkshopByID(id)
	if err != nil {
		return ds.Workshop{}, err
	}
	workshop.Name = req.Name
	workshop.Description = req.Description
	workshop.Century = req.Century
	if err := r.db.Save(&workshop).Error; err != nil {
		return ds.Workshop{}, err
	}
	return workshop, nil
}

func (r *Repository) DeleteWorkshop(id uint) error {
	workshop, err := r.GetWorkshopByID(id)
	if err != nil {
		return err
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workshop_id = ?", id).Delete(&ds.WorkshopProduction{}).Error; err != nil {
			return err
		}
		ctx := context.Background()
		if err := r.mc.DeleteImage(ctx, workshop.ImageKey.String); err != nil {
			fmt.Printf("failed to delete main image from minio: %v\n", err)
		}
		if err := r.mc.DeleteImage(ctx, workshop.ExtraImageKey.String); err != nil {
			fmt.Printf("failed to delete extra image from minio: %v\n", err)
		}
		if err := tx.Delete(&ds.Workshop{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) UploadWorkshopImage(id uint, image, extraImage *multipart.FileHeader) (ds.Workshop, error) {
	workshop, err := r.GetWorkshopByID(id)
	if err != nil {
		return ds.Workshop{}, err
	}
	ctx := context.Background()
	if image != nil {
		if err := r.mc.DeleteImage(ctx, workshop.ImageKey.String); err != nil {
			return ds.Workshop{}, fmt.Errorf("failed to delete old main image: %w", err)
		}
		imageKey, err := r.mc.UploadImage(ctx, image)
		if err != nil {
			return ds.Workshop{}, fmt.Errorf("failed to upload main image: %w", err)
		}
		workshop.ImageKey = sql.NullString{String: imageKey, Valid: true}
	}
	if extraImage != nil {
		if err := r.mc.DeleteImage(ctx, workshop.ExtraImageKey.String); err != nil {
			return ds.Workshop{}, fmt.Errorf("failed to delete old extra image: %w", err)
		}
		extraImageKey, err := r.mc.UploadImage(ctx, extraImage)
		if err != nil {
			return ds.Workshop{}, fmt.Errorf("failed to upload extra image: %w", err)
		}
		workshop.ExtraImageKey = sql.NullString{String: extraImageKey, Valid: true}
	}
	if err := r.db.Save(&workshop).Error; err != nil {
		if workshop.ImageKey.Valid {
			_ = r.mc.DeleteImage(ctx, workshop.ImageKey.String)
		}
		if workshop.ExtraImageKey.Valid {
			_ = r.mc.DeleteImage(ctx, workshop.ExtraImageKey.String)
		}
		return ds.Workshop{}, err
	}
	return workshop, nil
}
