package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

// Device - модель услуги (режима движения с круиз-контролем)
type Device struct {
	ID          int
	Title       string
	Power       float64 // Базовый расход топлива (л/100км)
	Photo       string  // Ключ изображения в Minio
	Description string
	ServiceType string // city, highway, mixed
}

func (r *Repository) GetDevices() ([]Device, error) {
	devices := []Device{
		{
			ID:          1,
			Title:       "Городской режим - Компактный",
			Power:       8.0,
			Photo:       "car_city_compact.jpg",
			Description: "Расчет экономии топлива для городского режима движения. Круиз-контроль в городе дает экономию около 5% за счет поддержания оптимальной скорости и плавного разгона.",
			ServiceType: "city",
		},
		{
			ID:          2,
			Title:       "Городской режим - Седан",
			Power:       10.0,
			Photo:       "car_city_sedan.jpg",
			Description: "Расчет экономии топлива для городского режима на седанах и кроссоверах. Круиз-контроль помогает избежать резких ускорений и торможений в пробках.",
			ServiceType: "city",
		},
		{
			ID:          3,
			Title:       "Трасса - Компактный",
			Power:       6.0,
			Photo:       "car_highway_compact.jpg",
			Description: "Расчет экономии топлива для трассы на компактных автомобилях. Круиз-контроль на трассе наиболее эффективен - экономия до 15% за счет поддержания постоянной скорости.",
			ServiceType: "highway",
		},
		{
			ID:          4,
			Title:       "Трасса - Внедорожник",
			Power:       12.0,
			Photo:       "car_highway_suv.jpg",
			Description: "Расчет экономии топлива для трассы на внедорожниках. На больших автомобилях круиз-контроль дает максимальную экономию за счет исключения человеческих ошибок.",
			ServiceType: "highway",
		},
		{
			ID:          5,
			Title:       "Смешанный режим - Седан",
			Power:       9.0,
			Photo:       "car_mixed_sedan.jpg",
			Description: "Расчет экономии топлива для смешанного режима движения (город + трасса). Круиз-контроль дает среднюю экономию около 10%. Идеально для междугородних поездок.",
			ServiceType: "mixed",
		},
		{
			ID:          6,
			Title:       "Смешанный режим - Грузовой",
			Power:       15.0,
			Photo:       "car_mixed_truck.jpg",
			Description: "Расчет экономии топлива для смешанного режима на грузовых автомобилях. Круиз-контроль критически важен для коммерческого транспорта - экономия до 12% на длинных маршрутах.",
			ServiceType: "mixed",
		},
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("Массив пустой")
	}
	return devices, nil
}

func (r *Repository) GetDevice(id int) (Device, error) {
	devices, err := r.GetDevices()
	if err != nil {
		return Device{}, err
	}
	for _, device := range devices {
		if device.ID == id {
			return device, nil
		}
	}
	return Device{}, fmt.Errorf("Услуга не найдена")
}

func (r *Repository) GetDeviceByTitle(title string) ([]Device, error) {
	devices, err := r.GetDevices()
	if err != nil {
		return []Device{}, err
	}
	var result []Device
	for _, device := range devices {
		if strings.Contains(strings.ToLower(device.Title), strings.ToLower(title)) {
			result = append(result, device)
		}
	}
	return result, nil
}

func (r *Repository) GetCart() ([]Device, error) {
	devices, err := r.GetDevices()
	if err != nil {
		return []Device{}, err
	}
	// Возвращаем 2 услуги для демонстрации заявки
	var result []Device
	for _, device := range devices {
		if device.ID == 3 || device.ID == 5 {
			result = append(result, device)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("Массив пустой")
	}
	return result, nil
}

// FilterDevices - фильтрация по полю (name, price, power)
func (r *Repository) FilterDevices(query, field string) ([]Device, error) {
	devices, err := r.GetDevices()
	if err != nil {
		return []Device{}, err
	}
	if query == "" {
		return devices, nil
	}
	var result []Device
	for _, device := range devices {
		match := false
		switch field {
		case "name":
			match = strings.Contains(strings.ToLower(device.Title), strings.ToLower(query))
		case "price":
			match = strings.Contains(fmt.Sprintf("%.1f", device.Power), query)
		case "power":
			match = strings.Contains(strings.ToLower(device.ServiceType), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(device.Title), strings.ToLower(query))
		}
		if match {
			result = append(result, device)
		}
	}
	return result, nil
}
