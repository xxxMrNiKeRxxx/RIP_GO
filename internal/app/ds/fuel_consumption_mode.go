package ds

type FuelConsumptionMode struct {
	ID            uint            `gorm:"primaryKey"`
	ConsumptionID uint            `gorm:"not null"`
	Consumption   FuelConsumption `gorm:"foreignKey:ConsumptionID"`
	ModeID        uint            `gorm:"not null"`
	Mode          DrivingMode     `gorm:"foreignKey:ModeID;references:ModeID"`
	RouteDistance int             `gorm:"not null"`
	FuelSaved     float64         `gorm:"default:0"`
	SortOrder     int             `gorm:"default:0"`
}

func (FuelConsumptionMode) TableName() string {
	return "fuel_consumption_modes"
}
