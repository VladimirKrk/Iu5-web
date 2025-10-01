package ds

type OrderWorkshop struct {
	ID              uint `gorm:"primaryKey"`
	ApplicationID   uint `gorm:"not null;uniqueIndex:idx_app_service"`
	ServiceID       uint `gorm:"not null;uniqueIndex:idx_app_service"`
	FoundDefects    int
	PredictedOutput string          `gorm:"type:varchar(255)"`
	Application     ProductionOrder `gorm:"foreignKey:ApplicationID"`
	Service         Workshop        `gorm:"foreignKey:ServiceID"`
}

// table stays the same
func (ow OrderWorkshop) TableName() string { return "application_services" }
