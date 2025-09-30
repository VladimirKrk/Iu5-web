package repository

import (
	"Iu5-web/internal/app/ds" // Убедитесь, что здесь Iu5-web
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ДОБАВЬТЕ ЭТУ СТРОКУ
const MINIO_URL = "http://localhost:9000/vlk-images/"

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// --------------------------------------------------------------------
// МЕТОДЫ ДЛЯ РАБОТЫ С УСЛУГАМИ (SERVICES)
// --------------------------------------------------------------------

// GetAllServices получает все активные услуги (через GORM)
func (r *Repository) GetAllServices() ([]ds.Service, error) {
	var services []ds.Service
	// Ищем все услуги, у которых is_deleted = false
	result := r.db.Where("is_deleted = ?", false).Find(&services)
	return services, result.Error
}

// GetServicesByName ищет активные услуги по имени (через GORM)
func (r *Repository) GetServicesByName(name string) ([]ds.Service, error) {
	var services []ds.Service
	// Ищем услуги, где имя содержит подстроку (ILIKE - регистронезависимый поиск)
	// и которые не удалены
	result := r.db.Where("name ILIKE ? AND is_deleted = ?", "%"+name+"%", false).Find(&services)
	return services, result.Error
}

// GetServiceByID находит одну услугу по ее ID (через GORM)
func (r *Repository) GetServiceByID(id uint) (ds.Service, error) {
	var service ds.Service
	result := r.db.First(&service, id)
	return service, result.Error
}

// --------------------------------------------------------------------
// МЕТОДЫ ДЛЯ РАБОТЫ С ЗАЯВКАМИ (APPLICATIONS)
// --------------------------------------------------------------------

// GetApplicationByID находит одну заявку по ее ID (через GORM)
// Preload("Items.Service") автоматически подгружает связанные данные!
func (r *Repository) GetApplicationByID(id uint) (ds.Application, error) {
	var application ds.Application
	// .Preload("Items.Service") - это "магия" GORM.
	// Он выполнит JOIN и заполнит данные по каждой услуге в заявке.
	result := r.db.Preload("Items.Service").First(&application, id)
	return application, result.Error
}

// FindOrCreateDraftApplication находит или создает заявку-черновик для пользователя (через GORM)
func (r *Repository) FindOrCreateDraftApplication(userID uint) (ds.Application, error) {
	var application ds.Application
	// Ищем заявку со статусом 'черновик' для данного пользователя
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&application).Error

	// Если заявка не найдена, создаем новую
	if err == gorm.ErrRecordNotFound {
		newApp := ds.Application{
			Status:    "черновик",
			CreatedAt: time.Now(),
			CreatorID: userID,
		}
		err = r.db.Create(&newApp).Error
		return newApp, err
	}

	return application, err
}

// AddServiceToApplication добавляет услугу в заявку (через GORM)
func (r *Repository) AddServiceToApplication(appID, serviceID uint) error {
	item := ds.ApplicationService{
		ApplicationID: appID,
		ServiceID:     serviceID,
		// Здесь можно будет добавить значения по умолчанию для found_defects и т.д., если нужно
	}
	// Создаем новую запись в таблице-связке
	result := r.db.Create(&item)
	return result.Error
}

// DeleteApplicationLogically выполняет логическое удаление заявки (через чистый SQL)
func (r *Repository) DeleteApplicationLogically(appID, userID uint) error {
	// Выполняем SQL UPDATE запрос, как того требует методичка.
	// Мы также проверяем, что пользователь может удалить только свою заявку.
	result := r.db.Exec("UPDATE applications SET status = ? WHERE id = ? AND creator_id = ?", "удалён", appID, userID)

	// Проверяем, что хотя бы одна строка была затронута
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Возвращаем ошибку, если заявка не найдена или не принадлежит пользователю
	}

	return result.Error
}

// GetDraftApplicationItemCount получает количество услуг в заявке-черновике пользователя
func (r *Repository) GetDraftApplicationItemCount(userID uint) (int64, error) {
	var application ds.Application
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&application).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil // Если черновика нет, то и товаров в нем 0. Это не ошибка.
		}
		return 0, err
	}

	var count int64
	err = r.db.Model(&ds.ApplicationService{}).Where("application_id = ?", application.ID).Count(&count).Error
	return count, err
}
