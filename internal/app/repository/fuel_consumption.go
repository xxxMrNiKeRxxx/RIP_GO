package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"web_backend/internal/app/ds"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	StatusDraft     = "draft"
	StatusDeleted   = "deleted"
	StatusFormed    = "formed"
	StatusCompleted = "completed"
	StatusRejected  = "rejected"
)

func (r *Repository) GetFuelConsumptionCount(creatorID uint) int64 {
	var appID uint
	err := r.db.Model(&ds.FuelConsumption{}).
		Where("creator_id = ? AND status = ?", creatorID, StatusDraft).
		Select("consumption_id").First(&appID).Error
	if err != nil {
		return 0
	}

	var count int64
	err = r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ?", appID).Count(&count).Error
	if err != nil {
		logrus.Error("Error counting fuel_consumption_modes:", err)
	}
	return count
}

func (r *Repository) GetActiveFuelConsumptionID(creatorID uint) uint {
	var appID uint
	err := r.db.Model(&ds.FuelConsumption{}).
		Where("creator_id = ? AND status = ?", creatorID, StatusDraft).
		Select("consumption_id").First(&appID).Error
	if err != nil {
		return 0
	}
	return appID
}

func (r *Repository) GetFuelConsumption(id int, creatorID uint) (*ds.FuelConsumption, []ds.FuelConsumptionMode, error) {
	var app ds.FuelConsumption
	err := r.db.Where("consumption_id = ? AND creator_id = ?", id, creatorID).First(&app).Error
	if err != nil {
		return nil, nil, err
	}

	var items []ds.FuelConsumptionMode
	err = r.db.Where("consumption_id = ?", id).
		Preload("Mode").
		Order("sort_order ASC, mode_id ASC").Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	return &app, items, nil
}

func (r *Repository) GetDraftConsumption(creatorID uint) (*ds.FuelConsumption, error) {
	var app ds.FuelConsumption
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, StatusDraft).
		First(&app).Error
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *Repository) AddModeToConsumption(modeID uint, creatorID uint, routeDistance int) error {
	var app ds.FuelConsumption

	err := r.db.Where("creator_id = ? AND status = ?", creatorID, StatusDraft).
		First(&app).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		app = ds.FuelConsumption{
			ApplicationID: fmt.Sprintf("APP-CC-%d", time.Now().Unix()%1000000), // ← Исправлено
			Status:        StatusDraft,
			CreatedAt:     time.Now(),
			CreatorID:     creatorID,
			Origin:        "Москва",
			Destination:   "Санкт-Петербург",
			FuelPrice:     55.00,
		}
		if err := r.db.Create(&app).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var count int64
	r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ? AND mode_id = ?", app.ConsumptionID, modeID).
		Count(&count)

	if count == 0 {
		var mode ds.DrivingMode
		if err := r.db.First(&mode, modeID).Error; err != nil {
			return err
		}

		var maxOrder int
		r.db.Model(&ds.FuelConsumptionMode{}).
			Where("consumption_id = ?", app.ConsumptionID).
			Select("COALESCE(MAX(sort_order), -1)").Scan(&maxOrder)

		fuelWithoutCruise := (mode.BaseConsumption / 100) * float64(routeDistance)
		fuelSaved := fuelWithoutCruise * (mode.EconomyPercent / 100)

		item := ds.FuelConsumptionMode{
			ConsumptionID: app.ConsumptionID,
			ModeID:        modeID,
			RouteDistance: routeDistance,
			FuelSaved:     fuelSaved,
			SortOrder:     maxOrder + 1,
		}
		if err := r.db.Create(&item).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) DeleteFuelConsumption(appID uint) error {
	query := `
		UPDATE fuel_consumptions 
		SET status = 'deleted'
		WHERE consumption_id = $1;
	`
	result := r.db.Exec(query, appID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("fuel_consumption with id %d not found", appID)
	}
	return nil
}

func (r *Repository) IsDraftFuelConsumption(appID int, creatorID uint) (bool, error) {
	var app ds.FuelConsumption
	err := r.db.Select("status").Where("consumption_id = ? AND creator_id = ?",
		appID, creatorID).First(&app).Error
	if err != nil {
		return false, err
	}
	return app.Status == StatusDraft, nil
}

func (r *Repository) CompleteFuelConsumption(appID uint) error {
	var totalSaved sql.NullFloat64
	err := r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ?", appID).
		Select("SUM(fuel_saved)").Scan(&totalSaved).Error
	if err != nil {
		return err
	}

	query := `
		UPDATE fuel_consumptions 
		SET total_saved = COALESCE($1, 0),
		    finish_date = CURRENT_TIMESTAMP,
		    status = 'completed'
		WHERE consumption_id = $2;
	`
	result := r.db.Exec(query, totalSaved.Float64, appID)
	return result.Error
}

func (r *Repository) UpdateRouteDistance(appID, modeID uint, routeDistance int) error {
	var mode ds.DrivingMode
	if err := r.db.First(&mode, modeID).Error; err != nil {
		return err
	}

	fuelWithoutCruise := (mode.BaseConsumption / 100) * float64(routeDistance)
	fuelSaved := fuelWithoutCruise * (mode.EconomyPercent / 100)

	if err := r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ? AND mode_id = ?", appID, modeID).
		Updates(map[string]interface{}{"route_distance": routeDistance, "fuel_saved": fuelSaved}).Error; err != nil {
		return err
	}

	return nil
}

// CalculateTotalSaved рассчитывает общую экономию для заявки
func (r *Repository) CalculateTotalSaved(appID uint) (float64, error) {
	var totalSaved sql.NullFloat64
	err := r.db.Model(&ds.FuelConsumptionMode{}).
		Where("consumption_id = ?", appID).
		Select("SUM(fuel_saved)").Scan(&totalSaved).Error
	if err != nil {
		return 0, err
	}
	return totalSaved.Float64, nil
}

// UpdateTotalSaved обновляет поле total_saved в заявке (ORM)
func (r *Repository) UpdateTotalSaved(appID uint, totalSaved float64) error {
	return r.db.Model(&ds.FuelConsumption{}).
		Where("consumption_id = ?", appID).
		Update("total_saved", totalSaved).Error
}

func (r *Repository) MoveModeInConsumption(appID, modeID uint, direction int) error {
	var items []ds.FuelConsumptionMode
	if err := r.db.Where("consumption_id = ?", appID).
		Order("sort_order ASC, mode_id ASC").Find(&items).Error; err != nil {
		return err
	}

	idx := -1
	for i := range items {
		if items[i].ModeID == modeID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("mode %d not in consumption %d", modeID, appID)
	}

	newIdx := idx + direction
	if newIdx < 0 || newIdx >= len(items) {
		return nil
	}

	items[idx], items[newIdx] = items[newIdx], items[idx]

	for i := range items {
		if err := r.db.Model(&ds.FuelConsumptionMode{}).
			Where("consumption_id = ? AND mode_id = ?", appID, items[i].ModeID).
			Update("sort_order", i).Error; err != nil {
			return err
		}
	}

	return nil
}
