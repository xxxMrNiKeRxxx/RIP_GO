package ds

type FuelConsumptionMode struct {
	ConsumptionID uint    `gorm:"primaryKey;column:consumption_id"`
	ModeID        uint    `gorm:"primaryKey;column:mode_id"`
	RouteDistance int     `gorm:"column:route_distance;not null"`
	FuelSaved     float64 `gorm:"column:fuel_saved;type:numeric(10,2);not null"`
	IsPrimary     *uint   `gorm:"column:is_primary"`
	SortOrder     int     `gorm:"column:sort_order;default:0"`

	Mode            DrivingMode     `gorm:"foreignKey:ModeID"`
	PrimaryMode     *DrivingMode    `gorm:"foreignKey:IsPrimary"`
	FuelConsumption FuelConsumption `gorm:"foreignKey:ConsumptionID"`
}

func (FuelConsumptionMode) TableName() string {
	return "fuel_consumption_modes"
}
