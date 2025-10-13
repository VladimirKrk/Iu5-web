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

// Методы для Мастерских (Workshops)
func (r *Repository) GetWorkshops(name string) ([]ds.Workshop, error) {
	var workshops []ds.Workshop
	query := r.db.Where("is_deleted = ?", false)
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	err := query.Find(&workshops).Error
	return workshops, err
}

func (r *Repository) GetWorkshopByID(id uint) (ds.Workshop, error) {
	var workshop ds.Workshop
	err := r.db.First(&workshop, id).Error
	return workshop, err
}

func (r *Repository) CreateWorkshop(workshop *ds.Workshop) error {
	return r.db.Create(workshop).Error
}

func (r *Repository) UpdateWorkshop(workshop *ds.Workshop) error {
	return r.db.Save(workshop).Error
}

func (r *Repository) DeleteWorkshop(id uint) error {
	return r.db.Delete(&ds.Workshop{}, id).Error
}

// Методы для Заказов (Applications/Orders)
func (r *Repository) GetApplications(status string, dateFrom, dateTo time.Time) ([]ds.WorkshopApplication, error) {
	var apps []ds.WorkshopApplication
	query := r.db.Where("status NOT IN (?, ?)", "черновик", "удалён")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if !dateFrom.IsZero() && !dateTo.IsZero() {
		query = query.Where("formed_at BETWEEN ? AND ?", dateFrom, dateTo)
	}
	err := query.Preload("Creator").Preload("Moderator").Find(&apps).Error
	return apps, err
}

func (r *Repository) GetApplicationByID(id uint) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	err := r.db.Preload("Items.Workshop").Preload("Creator").Preload("Moderator").First(&app, id).Error
	return app, err
}

func (r *Repository) FindOrCreateDraftApplication(userID uint) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&app).Error
	if err == gorm.ErrRecordNotFound {
		newApp := ds.WorkshopApplication{Status: "черновик", CreatedAt: time.Now(), CreatorID: userID}
		err = r.db.Create(&newApp).Error
		return newApp, err
	}
	return app, err
}

func (r *Repository) UpdateApplication(app *ds.WorkshopApplication) error {
	return r.db.Save(app).Error
}

func (r *Repository) DeleteApplicationLogically(appID, userID uint) error {
	result := r.db.Exec("UPDATE workshop_applications SET status = ? WHERE id = ? AND creator_id = ?", "удалён", appID, userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// Методы для Позиций Заказа (Order Items)
func (r *Repository) AddWorkshopToApplication(appID, workshopID uint) (ds.WorkshopProduction, error) {
	item := ds.WorkshopProduction{ApplicationID: appID, WorkshopID: workshopID}
	err := r.db.Create(&item).Error
	return item, err
}

func (r *Repository) GetApplicationItem(appID, workshopID uint) (ds.WorkshopProduction, error) {
	var item ds.WorkshopProduction
	err := r.db.Where("application_id = ? AND workshop_id = ?", appID, workshopID).First(&item).Error
	return item, err
}

func (r *Repository) UpdateApplicationItem(item *ds.WorkshopProduction) error {
	return r.db.Save(item).Error
}

func (r *Repository) DeleteApplicationItem(appID, workshopID uint) error {
	result := r.db.Where("application_id = ? AND workshop_id = ?", appID, workshopID).Delete(&ds.WorkshopProduction{})
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

// Методы для Пользователей (Users)
func (r *Repository) CreateUser(user *ds.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	err := r.db.First(&user, id).Error
	return user, err
}
