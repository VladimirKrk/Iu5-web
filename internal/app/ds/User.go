package ds

// User представляет таблицу users
type User struct {
	ID          uint   `gorm:"primaryKey"`
	Login       string `gorm:"type:varchar(50);unique;not null"`
	Password    string `gorm:"type:varchar(255);not null"`
	IsModerator bool   `gorm:"default:false"`
}

// Указываем GORM, что эта структура соответствует таблице 'users'
func (u User) TableName() string {
	return "users"
}
