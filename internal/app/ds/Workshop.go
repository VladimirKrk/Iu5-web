package ds

import "database/sql"

// Workshop представляет сущность "Мастерская"
type Workshop struct {
	ID            uint           `gorm:"primaryKey"`
	Name          string         `gorm:"type:varchar(255);not null"`
	Description   string         `gorm:"type:text;not null"`
	Century       string         `gorm:"type:varchar(50)"`
	ImageKey      sql.NullString `gorm:"type:varchar(255)"`
	ExtraImageKey sql.NullString `gorm:"type:varchar(255)"`
}

func (w Workshop) TableName() string {
	return "workshops"
}
