package repository

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Repository struct{}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type DrivingMode struct {
	ID              int
	ModeName        string
	Description     string
	ImageKey        string
	VideoKey        string
	BaseConsumption float64
	EconomyPercent  float64
	DrivingType     string
}

type FuelConsumption struct {
	ApplicationID     string
	CreatedDate       string
	Status            string
	Origin            string
	Destination       string
	Modes             []FuelConsumptionMode
	RoutesCount       int
	CalculationResult string
}

type FuelConsumptionMode struct {
	Mode              DrivingMode
	RouteDistance     int
	FuelWithoutCruise float64
	FuelSaved         float64
}

func (r *Repository) GetDrivingModes() ([]DrivingMode, error) {
	modes := []DrivingMode{
		{
			ID:              1,
			ModeName:        "Городской режим - Компактный",
			Description:     "Расчет экономии топлива для городского режима движения.",
			ImageKey:        "car_city_compact.jpg",
			VideoKey:        "car_city_compact.mp4",
			BaseConsumption: 8.0,
			EconomyPercent:  5.0,
			DrivingType:     "city",
		},
		{
			ID:              2,
			ModeName:        "Городской режим - Седан",
			Description:     "Расчет экономии топлива для городского режима на седанах.",
			ImageKey:        "car_city_sedan.jpg",
			VideoKey:        "car_city_sedan.mp4",
			BaseConsumption: 10.0,
			EconomyPercent:  5.0,
			DrivingType:     "city",
		},
		{
			ID:              3,
			ModeName:        "Трасса - Компактный",
			Description:     "Расчет экономии топлива для трассы на компактных автомобилях.",
			ImageKey:        "car_highway_compact.jpg",
			VideoKey:        "car_highway_compact.mp4",
			BaseConsumption: 6.0,
			EconomyPercent:  15.0,
			DrivingType:     "highway",
		},
		{
			ID:              4,
			ModeName:        "Трасса - Внедорожник",
			Description:     "Расчет экономии топлива для трассы на внедорожниках.",
			ImageKey:        "car_highway_suv.jpg",
			VideoKey:        "car_highway_suv.mp4",
			BaseConsumption: 12.0,
			EconomyPercent:  15.0,
			DrivingType:     "highway",
		},
		{
			ID:              5,
			ModeName:        "Смешанный режим - Седан",
			Description:     "Расчет экономии топлива для смешанного режима движения.",
			ImageKey:        "car_mixed_sedan.jpg",
			VideoKey:        "car_mixed_sedan.mp4",
			BaseConsumption: 9.0,
			EconomyPercent:  10.0,
			DrivingType:     "mixed",
		},
		{
			ID:              6,
			ModeName:        "Смешанный режим - Грузовой",
			Description:     "Расчет экономии топлива для смешанного режима на грузовых автомобилях.",
			ImageKey:        "car_mixed_truck.jpg",
			VideoKey:        "car_mixed_truck.mp4",
			BaseConsumption: 15.0,
			EconomyPercent:  10.0,
			DrivingType:     "mixed",
		},
	}
	return modes, nil
}

func (r *Repository) GetDrivingMode(id int) (DrivingMode, error) {
	modes, _ := r.GetDrivingModes()
	for _, m := range modes {
		if m.ID == id {
			return m, nil
		}
	}
	return DrivingMode{}, fmt.Errorf("режим не найден")
}

func (r *Repository) GetDrivingModesBySearch(searchText, searchField string) ([]DrivingMode, error) {
	modes, _ := r.GetDrivingModes()
	if searchText == "" {
		return modes, nil
	}
	var result []DrivingMode
	for _, m := range modes {
		match := false
		if searchField == "name" {
			match = strings.Contains(strings.ToLower(m.ModeName), strings.ToLower(searchText))
		} else if searchField == "consumption" {
			searchNum, err := strconv.ParseFloat(searchText, 64)
			if err == nil {
				match = math.Abs(m.BaseConsumption-searchNum) < 0.1
			}
		}
		if match {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *Repository) GetFuelConsumptions() ([]FuelConsumption, error) {
	entries := []struct {
		ModeID        int
		RouteDistance int
	}{
		{3, 300},
		{5, 500},
	}

	modes, _ := r.GetDrivingModes()
	modeMap := make(map[int]DrivingMode)
	for _, m := range modes {
		modeMap[m.ID] = m
	}

	var consumptionModes []FuelConsumptionMode
	for _, e := range entries {
		mode, ok := modeMap[e.ModeID]
		if !ok {
			continue
		}
		fuelWithoutCruise := (mode.BaseConsumption / 100) * float64(e.RouteDistance)
		fuelSaved := fuelWithoutCruise * (mode.EconomyPercent / 100)
		consumptionModes = append(consumptionModes, FuelConsumptionMode{
			Mode:              mode,
			RouteDistance:     e.RouteDistance,
			FuelWithoutCruise: fuelWithoutCruise,
			FuelSaved:         fuelSaved,
		})
	}

	totalSaved := 0.0
	for _, cm := range consumptionModes {
		totalSaved += cm.FuelSaved
	}

	consumption := FuelConsumption{
		ApplicationID:     "APP-CC-2026-001",
		CreatedDate:       "25.01.2026",
		Status:            "calculated",
		Origin:            "Москва",
		Destination:       "Санкт-Петербург",
		Modes:             consumptionModes,
		RoutesCount:       len(consumptionModes),
		CalculationResult: fmt.Sprintf("Экономия: %.2f л | %.0f ₽", totalSaved, totalSaved*55),
	}

	return []FuelConsumption{consumption}, nil
}

func (r *Repository) GetFuelConsumption(id int) (FuelConsumption, error) {
	consumptions, _ := r.GetFuelConsumptions()
	if len(consumptions) > 0 {
		return consumptions[0], nil
	}
	return FuelConsumption{}, fmt.Errorf("заявка не найдена")
}

func (r *Repository) GetFuelConsumptionForMode(modeID int) (*FuelConsumptionMode, error) {
	consumptions, _ := r.GetFuelConsumptions()
	for _, c := range consumptions {
		for _, cm := range c.Modes {
			if cm.Mode.ID == modeID {
				return &cm, nil
			}
		}
	}
	return nil, fmt.Errorf("режим не найден в заявках")
}
