package ds

type DrivingMode struct {
	ModeID          uint    `gorm:"primaryKey"`
	ModeName        string  `gorm:"type:varchar(100);not null"`
	Description     string  `gorm:"type:text"`
	ImageKey        string  `gorm:"type:varchar(255)"`
	VideoKey        string  `gorm:"type:varchar(255)"`
	BaseConsumption float64 `gorm:"not null"`
	EconomyPercent  float64 `gorm:"not null"`
	DrivingType     string  `gorm:"type:varchar(20);not null"`
	IsActive        bool    `gorm:"type:boolean;not null;default:false"`
}
