package serializer

import "metoda/internal/app/ds"

type ModeJSON struct {
	ModeID          uint    `json:"mode_id"`
	ModeName        string  `json:"mode_name"`
	Description     string  `json:"description"`
	ImageKey        string  `json:"image_key"`
	VideoKey        string  `json:"video_key"`
	BaseConsumption float64 `json:"base_consumption"`
	EconomyPercent  float64 `json:"economy_percent"`
	DrivingType     string  `json:"driving_type"`
	IsActive        bool    `json:"is_active"`
}

func ModeToJSON(t ds.DrivingMode) ModeJSON {
	return ModeJSON{
		ModeID:          t.ModeID,
		ModeName:        t.ModeName,
		Description:     t.Description,
		ImageKey:        t.ImageKey,
		VideoKey:        t.VideoKey,
		BaseConsumption: t.BaseConsumption,
		EconomyPercent:  t.EconomyPercent,
		DrivingType:     t.DrivingType,
		IsActive:        t.IsActive,
	}
}

func ModeFromJSON(j ModeJSON) ds.DrivingMode {
	return ds.DrivingMode{
		ModeName:        j.ModeName,
		Description:     j.Description,
		ImageKey:        j.ImageKey,
		VideoKey:        j.VideoKey,
		BaseConsumption: j.BaseConsumption,
		EconomyPercent:  j.EconomyPercent,
		DrivingType:     j.DrivingType,
		IsActive:        j.IsActive,
	}
}
