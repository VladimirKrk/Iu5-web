package ds

type WorkshopProduction struct {
	ID              uint `gorm:"primaryKey"`
	ApplicationID   uint `gorm:"not null;uniqueIndex:idx_app_workshop"`
	WorkshopID      uint `gorm:"not null;uniqueIndex:idx_app_workshop"`
	FoundDefects    int
	PredictedOutput string              `gorm:"type:varchar(255)"`
	Application     WorkshopApplication `gorm:"foreignKey:ApplicationID"`
	Workshop        Workshop            `gorm:"foreignKey:WorkshopID"`
}

func (wp WorkshopProduction) TableName() string { return "workshop_production" }
