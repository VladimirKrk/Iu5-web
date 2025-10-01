package ds

type Workshop struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"type:varchar(255);not null"`
	Description   string `gorm:"type:text"`
	Century       string `gorm:"type:varchar(50)"`
	ImageKey      string `gorm:"type:varchar(255)"`
	ExtraImageKey string `gorm:"type:varchar(255)"`
	IsDeleted     bool   `gorm:"default:false"`
}

func (w Workshop) TableName() string { return "services" } // Таблица остается services
