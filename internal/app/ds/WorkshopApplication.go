// FILE: internal/app/ds/WorkshopApplication.go
package ds

import (
	"database/sql"
	"time"
)

// WorkshopApplication представляет заявку на расчет производства
type WorkshopApplication struct {
	ID             uint           `gorm:"primaryKey"`
	Status         string         `gorm:"type:varchar(50);not null"` // статусы: draft, deleted, formed, completed
	ProductionName sql.NullString `gorm:"type:varchar(255)"`
	CreatedAt      time.Time      `gorm:"not null"`
	FormedAt       sql.NullTime   `gorm:"default:null"`
	CompletedAt    sql.NullTime   `gorm:"default:null"`
	CreatorID      uint           `gorm:"not null"`
	ModeratorID    sql.NullInt64  `gorm:"default:null"`
	Creator        User           `gorm:"foreignKey:CreatorID"`
	Moderator      User           `gorm:"foreignKey:ModeratorID"`
}

func (wa WorkshopApplication) TableName() string {
	return "workshop_applications"
}
