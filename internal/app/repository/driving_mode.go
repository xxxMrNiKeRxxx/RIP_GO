package repository

import (
	"fmt"
	"web_backend/internal/app/ds"
)

func (r *Repository) GetDrivingModes() ([]ds.DrivingMode, error) {
	var modes []ds.DrivingMode
	err := r.db.Where("is_active = ?", true).Find(&modes).Error
	if err != nil {
		return nil, err
	}
	if len(modes) == 0 {
		return nil, fmt.Errorf("массив режимов пуст")
	}
	return modes, nil
}

func (r *Repository) GetDrivingMode(id int) (ds.DrivingMode, error) {
	var mode ds.DrivingMode
	err := r.db.Where("mode_id = ? AND is_active = ?", id, true).First(&mode).Error
	if err != nil {
		return ds.DrivingMode{}, err
	}
	return mode, nil
}

func (r *Repository) GetDrivingModesBySearch(searchText string) ([]ds.DrivingMode, error) {
	var modes []ds.DrivingMode
	err := r.db.Where("mode_name ILIKE ? AND is_active = ?", "%"+searchText+"%", true).Find(&modes).Error
	if err != nil {
		return nil, err
	}
	return modes, nil
}

func (r *Repository) GetDrivingModesByConsumption(consumption float64) ([]ds.DrivingMode, error) {
	var modes []ds.DrivingMode
	err := r.db.Where("base_consumption = ? AND is_active = ?", consumption, true).Find(&modes).Error
	if err != nil {
		return nil, err
	}
	return modes, nil
}
