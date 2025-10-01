package ds

import "time"

type ProductionOrder struct { // Application -> ProductionOrder
	ID          uint      `gorm:"primaryKey"`
	Status      string    `gorm:"type:varchar(50);not null"`
	CreatedAt   time.Time `gorm:"not null"`
	CreatorID   uint      `gorm:"not null"`
	FormedAt    *time.Time
	CompletedAt *time.Time
	ModeratorID *uint
	Creator     User            `gorm:"foreignKey:CreatorID"`
	Moderator   User            `gorm:"foreignKey:ModeratorID"`
	Items       []OrderWorkshop `gorm:"foreignKey:ApplicationID"` // ApplicationService -> OrderWorkshop
}

func (p ProductionOrder) TableName() string { return "applications" } // Таблица остается applications
