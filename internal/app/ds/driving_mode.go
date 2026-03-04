package ds

type DrivingMode struct {
	ModeID          uint    `gorm:"primaryKey;column:mode_id"`
	ModeName        string  `gorm:"type:varchar(100);not null"`
	Description     string  `gorm:"type:text;not null"`
	IsActive        bool    `gorm:"type:boolean;not null;default:true"`
	ImageKey        string  `gorm:"column:image_key;type:varchar(255)"`
	VideoKey        string  `gorm:"column:video_key;type:varchar(255)"`
	BaseConsumption float64 `gorm:"type:numeric(5,2);not null"`
	EconomyPercent  float64 `gorm:"type:numeric(5,2);not null"`
	DrivingType     string  `gorm:"type:varchar(20);not null"`
}

func (DrivingMode) TableName() string {
	return "driving_modes"
}
