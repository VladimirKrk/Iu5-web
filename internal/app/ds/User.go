package ds

type User struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Login       string `gorm:"type:varchar(50);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(255);not null" json:"password"`
	IsModerator bool   `gorm:"default:false" json:"is_moderator"`
}

func (User) TableName() string { return "users" }
