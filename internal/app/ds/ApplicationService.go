package ds

// ApplicationService представляет таблицу application_services
type ApplicationService struct {
	ID              uint `gorm:"primaryKey"`
	ApplicationID   uint `gorm:"not null;uniqueIndex:idx_app_service"`
	ServiceID       uint `gorm:"not null;uniqueIndex:idx_app_service"`
	FoundDefects    int
	PredictedOutput string `gorm:"type:varchar(255)"`

	// Связи
	Application Application `gorm:"foreignKey:ApplicationID"`
	Service     Service     `gorm:"foreignKey:ServiceID"`
}

func (as ApplicationService) TableName() string {
	return "application_services"
}
