package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"metoda/internal/app/ds"
	"metoda/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Мигрируем в правильном порядке
	err = db.AutoMigrate(
		&ds.Users{},               // 1. Без внешних ключей
		&ds.DrivingMode{},         // 2. Без внешних ключей
		&ds.FuelConsumption{},     // 3. Без внешних ключей
		&ds.FuelConsumptionMode{}, // 4. С внешними ключами (последним!)
	)
	if err != nil {
		panic("cant migrate db")
	}
}
