package ds

// Service представляет таблицу services
type Service struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"type:varchar(255);not null"`
	Description   string `gorm:"type:text"`
	Century       string `gorm:"type:varchar(50)"`
	ImageKey      string `gorm:"type:varchar(255)"`
	ExtraImageKey string `gorm:"type:varchar(255)"`
	IsDeleted     bool   `gorm:"default:false"`
}

func (s Service) TableName() string {
	return "services"
}
