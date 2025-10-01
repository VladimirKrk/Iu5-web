// internal/app/repository/repository.go
package repository

import (
	"Iu5-web/internal/app/ds"
	"log" // <-- ДОБАВЬТЕ
	"os"  // <-- ДОБАВЬТЕ
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger" // <-- ДОБАВЬТЕ
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
			LogLevel:      logger.Info, // Показываем все SQL-запросы
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger, // <-- ПРИМЕНЯЕМ ЛОГГЕР
	})
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// методы

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

// === МЕТОДЫ ДЛЯ ЗАКАЗОВ (PRODUCTION ORDERS) ===

func (r *Repository) GetOrderByID(id uint) (ds.ProductionOrder, error) {
	var order ds.ProductionOrder
	result := r.db.Preload("Items.Service").First(&order, id)
	return order, result.Error
}

func (r *Repository) FindOrCreateDraftOrder(userID uint) (ds.ProductionOrder, error) {
	var order ds.ProductionOrder
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&order).Error
	if err == gorm.ErrRecordNotFound {
		newOrder := ds.ProductionOrder{
			Status:    "черновик",
			CreatedAt: time.Now(),
			CreatorID: userID,
		}
		err = r.db.Create(&newOrder).Error
		return newOrder, err
	}
	return order, err
}

func (r *Repository) AddWorkshopToOrder(orderID, workshopID uint) error {
	item := ds.OrderWorkshop{
		ApplicationID: orderID,
		ServiceID:     workshopID,
	}
	result := r.db.Create(&item)
	return result.Error
}

func (r *Repository) DeleteOrderLogically(orderID, userID uint) error {
	result := r.db.Exec("UPDATE applications SET status = ? WHERE id = ? AND creator_id = ?", "удалён", orderID, userID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *Repository) GetDraftOrderItemCount(userID uint) (int64, error) {
	var order ds.ProductionOrder
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	var count int64
	err = r.db.Model(&ds.OrderWorkshop{}).Where("application_id = ?", order.ID).Count(&count).Error
	return count, err
}
