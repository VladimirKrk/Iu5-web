package ds

import "time"

// Application представляет таблицу applications
type Application struct {
	ID          uint      `gorm:"primaryKey"`
	Status      string    `gorm:"type:varchar(50);not null"`
	CreatedAt   time.Time `gorm:"not null"`
	CreatorID   uint      `gorm:"not null"`
	FormedAt    *time.Time
	CompletedAt *time.Time
	ModeratorID *uint

	// Связи (GORM будет использовать их для JOIN-запросов)
	Creator   User `gorm:"foreignKey:CreatorID"`
	Moderator User `gorm:"foreignKey:ModeratorID"`
	// 1 заявка имеет много Аппликатион Сервис
	Items []ApplicationService `gorm:"foreignKey:ApplicationID"`
}

func (a Application) TableName() string {
	return "applications"
}
