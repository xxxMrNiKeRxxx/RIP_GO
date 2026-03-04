package ds

import (
	"database/sql"
	"time"
)

type FuelConsumption struct {
	ConsumptionID uint         `gorm:"primaryKey;column:consumption_id"`
	ApplicationID string       `gorm:"type:varchar(20);not null;unique"`
	Status        string       `gorm:"type:varchar(20);not null"`
	CreatedAt     time.Time    `gorm:"not null"`
	CreatorID     uint         `gorm:"not null"`
	Origin        string       `gorm:"type:varchar(100);not null"`
	Destination   string       `gorm:"type:varchar(100);not null"`
	FormingDate   *time.Time   `gorm:"column:forming_date"`
	FinishDate    sql.NullTime `gorm:"column:finish_date"`
	ModeratorID   *uint        `gorm:"column:moderator_id"`
	FuelPrice     float64      `gorm:"type:numeric(10,2);default:55.00"`
	TotalSaved    *float64     `gorm:"column:total_saved;type:numeric(10,2)"`

	Creator   Users  `gorm:"foreignKey:CreatorID"`
	Moderator *Users `gorm:"foreignKey:ModeratorID"`
}

func (FuelConsumption) TableName() string {
	return "fuel_consumptions"
}
