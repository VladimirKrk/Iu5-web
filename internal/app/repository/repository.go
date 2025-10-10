package repository

import (
	"Iu5-web/internal/app/ds"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const MINIO_URL = "http://localhost:9000/vlk-images/"

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

func (r *Repository) GetAllWorkshops() ([]ds.Workshop, error) {
	var workshops []ds.Workshop
	result := r.db.Where("is_deleted = ?", false).Find(&workshops)
	return workshops, result.Error
}

func (r *Repository) GetWorkshopsByName(name string) ([]ds.Workshop, error) {
	var workshops []ds.Workshop
	result := r.db.Where("name ILIKE ? AND is_deleted = ?", "%"+name+"%", false).Find(&workshops)
	return workshops, result.Error
}

func (r *Repository) GetWorkshopByID(id uint) (ds.Workshop, error) {
	var workshop ds.Workshop
	result := r.db.First(&workshop, id)
	return workshop, result.Error
}

func (r *Repository) GetApplicationByID(id uint) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	result := r.db.Preload("Items.Workshop").First(&app, id)
	return app, result.Error
}

func (r *Repository) FindOrCreateDraftApplication(userID uint) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&app).Error
	if err == gorm.ErrRecordNotFound {
		newApp := ds.WorkshopApplication{
			Status:    "черновик",
			CreatedAt: time.Now(),
			CreatorID: userID,
		}
		err = r.db.Create(&newApp).Error
		return newApp, err
	}
	return app, err
}

func (r *Repository) AddWorkshopToApplication(appID, workshopID uint) error {
	item := ds.WorkshopProduction{
		ApplicationID: appID,
		WorkshopID:    workshopID,
	}
	result := r.db.Create(&item)
	return result.Error
}

func (r *Repository) DeleteApplicationLogically(appID, userID uint) error {
	result := r.db.Exec("UPDATE workshop_applications SET status = ? WHERE id = ? AND creator_id = ?", "удалён", appID, userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *Repository) GetDraftApplicationItemCount(userID uint) (int64, error) {
	var app ds.WorkshopApplication
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&app).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	var count int64
	err = r.db.Model(&ds.WorkshopProduction{}).Where("application_id = ?", app.ID).Count(&count).Error
	return count, err
}
