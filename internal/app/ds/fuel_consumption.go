package ds

import (
	"database/sql"
	"time"
)

const (
	StatusDraft     = "черновик"
	StatusDeleted   = "удалён"
	StatusFormed    = "сформирован"
	StatusCompleted = "завершён"
	StatusRejected  = "отклонён"
)

type FuelConsumption struct {
	ConsumptionID uint         `gorm:"primaryKey"`
	Status        string       `gorm:"type:varchar(20);not null"`
	DateCreate    time.Time    `gorm:"not null"`
	DateFormed    sql.NullTime `gorm:"default:null"`
	DateCompleted sql.NullTime `gorm:"default:null"`
	CreatorID     uint         `gorm:"not null"`
	ModeratorID   *uint        `gorm:"default:null"`
	Origin        string       `gorm:"default:null"`
	Destination   string       `gorm:"default:null"`
	FuelPrice     float64      `gorm:"default:55.00"`
	TotalSaved    float64      `gorm:"default:0"`
	Creator       Users        `gorm:"foreignKey:CreatorID"`
	Moderator     *Users       `gorm:"foreignKey:ModeratorID"`
}

func (FuelConsumption) TableName() string {
	return "fuel_consumptions"
}
