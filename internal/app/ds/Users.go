package ds

type Users struct {
	UserID      uint   `gorm:"primaryKey;column:user_id"`
	Login       string `gorm:"type:varchar(50);unique;not null"`
	Password    string `gorm:"type:varchar(100);not null"`
	IsModerator bool   `gorm:"type:boolean;default:false"`
}

func (Users) TableName() string {
	return "users"
}
