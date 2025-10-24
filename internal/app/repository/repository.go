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

// ==========================================
// Домен "Мастерские" (Workshops)
// ==========================================

func (r *Repository) GetWorkshops(name string) ([]ds.Workshop, error) {
	var workshops []ds.Workshop
	query := r.db.Where("is_deleted = ?", false)
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	return workshops, query.Find(&workshops).Error
}

func (r *Repository) GetWorkshopByID(id uint) (ds.Workshop, error) {
	var workshop ds.Workshop
	return workshop, r.db.First(&workshop, id).Error
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

// ==========================================
// Домен "Заявки" (Workshop Applications)
// ==========================================

func (r *Repository) GetWorkshopApplications(status string, dateFrom, dateTo time.Time) ([]ds.WorkshopApplication, error) {
	var apps []ds.WorkshopApplication
	query := r.db.Where("status NOT IN (?, ?)", "черновик", "удалён")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if !dateFrom.IsZero() && !dateTo.IsZero() {
		query = query.Where("formed_at BETWEEN ? AND ?", dateFrom, dateTo)
	}
	return apps, query.Preload("Creator").Preload("Moderator").Find(&apps).Error
}

func (r *Repository) GetWorkshopApplicationByID(id uint) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	return app, r.db.Preload("Items.Workshop").Preload("Creator").Preload("Moderator").First(&app, id).Error
}

func (r *Repository) FindOrCreateDraftApplication(userID uint) (ds.WorkshopApplication, error) {
	var app ds.WorkshopApplication
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&app).Error
	if err == gorm.ErrRecordNotFound {
		newApp := ds.WorkshopApplication{Status: "черновик", CreatedAt: time.Now(), CreatorID: userID}
		return newApp, r.db.Create(&newApp).Error
	}
	return app, err
}

func (r *Repository) UpdateWorkshopApplication(app *ds.WorkshopApplication) error {
	return r.db.Save(app).Error
}

func (r *Repository) DeleteApplicationLogically(appID, userID uint) error {
	result := r.db.Exec("UPDATE workshop_applications SET status = ? WHERE id = ? AND creator_id = ?", "удалён", appID, userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// ==========================================
// Домен "Продукция в заявке" (Workshop Production)
// ==========================================

func (r *Repository) AddWorkshopToApplication(appID, workshopID uint) (ds.WorkshopProduction, error) {
	item := ds.WorkshopProduction{ApplicationID: appID, WorkshopID: workshopID}
	return item, r.db.Create(&item).Error
}

func (r *Repository) GetWorkshopProductionItem(appID, workshopID uint) (ds.WorkshopProduction, error) {
	var item ds.WorkshopProduction
	return item, r.db.Where("application_id = ? AND workshop_id = ?", appID, workshopID).First(&item).Error
}

func (r *Repository) UpdateWorkshopProductionItem(item *ds.WorkshopProduction) error {
	return r.db.Save(item).Error
}

func (r *Repository) DeleteWorkshopProductionItem(appID, workshopID uint) error {
	return r.db.Where("application_id = ? AND workshop_id = ?", appID, workshopID).Delete(&ds.WorkshopProduction{}).Error
}

func (r *Repository) GetDraftItemCount(userID uint) (int64, error) {
	var app ds.WorkshopApplication
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&app).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	var count int64
	return count, r.db.Model(&ds.WorkshopProduction{}).Where("application_id = ?", app.ID).Count(&count).Error
}

// ==========================================
// Домен "Пользователи" (Users)
// ==========================================

func (r *Repository) CreateUser(user *ds.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	return user, r.db.First(&user, id).Error
}

func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
	var user ds.User
	return user, r.db.Where("login = ?", login).First(&user).Error
}

func (r *Repository) UpdateUser(user *ds.User) error {
	return r.db.Save(user).Error
}
