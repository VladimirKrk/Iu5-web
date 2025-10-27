// FILE: internal/app/ds/WorkshopProduction.go
package ds

import "database/sql"

// WorkshopProduction представляет связь "мастерская в заявке"
type WorkshopProduction struct {
	ApplicationID uint `gorm:"primaryKey;autoIncrement:false"`
	WorkshopID    uint `gorm:"primaryKey;autoIncrement:false"`

	// Дополнительные поля для этой связи
	FoundDefects    int            `gorm:"not null;default:0"`
	PredictedOutput sql.NullString `gorm:"type:varchar(255)"`

	// Связи с моделями
	Application WorkshopApplication `gorm:"foreignKey:ApplicationID"`
	Workshop    Workshop            `gorm:"foreignKey:WorkshopID"`
}

func (wp WorkshopProduction) TableName() string {
	return "workshop_production"
}
