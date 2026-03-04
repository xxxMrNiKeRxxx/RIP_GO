package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"web_backend/internal/app/ds"
	"web_backend/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.Users{},
		&ds.DrivingMode{},
		&ds.FuelConsumption{},
		&ds.FuelConsumptionMode{},
	)
	if err != nil {
		panic("cant migrate db")
	}
}
