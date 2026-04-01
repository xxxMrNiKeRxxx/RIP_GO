package serializer

import (
	"metoda/internal/app/ds"
	"time"
)

type FuelConsumptionListJSON struct {
	ConsumptionID    uint       `json:"consumption_id"`
	Status           string     `json:"status"`
	Created_at       time.Time  `json:"date_create"`
	DateFormed       *time.Time `json:"date_formed"`
	DateCompleted    *time.Time `json:"date_completed"`
	CreatorLogin     string     `json:"creator_login"`
	ModeratorLogin   *string    `json:"moderator_login"`
	Origin           string     `json:"origin"`
	Destination      string     `json:"destination"`
	FuelPrice        float64    `json:"fuel_price"`
	TotalSaved       float64    `json:"total_saved"` // ✅ Остался float64 (не *float64)
	FuelEntriesCount int        `json:"fuel_entries_count"`
}

type FuelConsumptionDetailJSON struct {
	ConsumptionID  uint                           `json:"consumption_id"`
	Status         string                         `json:"status"`
	Created_at     time.Time                      `json:"date_create"`
	DateFormed     *time.Time                     `json:"date_formed"`
	DateCompleted  *time.Time                     `json:"date_completed"`
	CreatorLogin   string                         `json:"creator_login"`
	ModeratorLogin *string                        `json:"moderator_login"`
	Origin         string                         `json:"origin"`
	Destination    string                         `json:"destination"`
	FuelPrice      float64                        `json:"fuel_price"`
	TotalSaved     float64                        `json:"total_saved"` // ✅ Остался float64
	Entries        []FuelConsumptionEntryViewJSON `json:"entries"`
}

// ✅ УБРАНО поле SortOrder
type FuelConsumptionEntryViewJSON struct {
	ID              uint    `json:"id"`
	ModeID          uint    `json:"mode_id"`
	ModeName        string  `json:"mode_name"`
	BaseConsumption float64 `json:"base_consumption"`
	EconomyPercent  float64 `json:"economy_percent"`
	ImageKey        string  `json:"image_key"`
	RouteDistance   int     `json:"route_distance"`
	FuelSaved       float64 `json:"fuel_saved"`

}

type FuelConsumptionUpdateJSON struct {
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	FuelPrice   float64 `json:"fuel_price"`
}

type FinishJSON struct {
	Status string `json:"status"`
}

type CartJSON struct {
	ConsumptionID uint  `json:"consumption_id"`
	ModesCount    int64 `json:"modes_count"`
}

// ✅ ИСПРАВЛЕНА функция: total_saved = 0, если не завершено
func FuelConsumptionToListJSON(t ds.FuelConsumption, creatorLogin string, moderatorLogin string, entriesCount int) FuelConsumptionListJSON {
	var dateFormed, dateCompleted *time.Time
	if t.DateFormed.Valid {
		dateFormed = &t.DateFormed.Time
	}
	if t.DateCompleted.Valid {
		dateCompleted = &t.DateCompleted.Time
	}

	var modLogin *string
	if moderatorLogin != "" {
		modLogin = &moderatorLogin
	}

	// ✅ Условие: fuel_entries_count = 0, если статус ≠ "завершён"
	fuelEntriesCount := 0
	if t.Status == ds.StatusCompleted {
		fuelEntriesCount = entriesCount // ✅ Правильно: присваивание
	}

	// ✅ Условие: total_saved = 0, если статус ≠ "завершён"
	totalSaved := float64(0)
	if t.Status == ds.StatusCompleted {
		totalSaved = t.TotalSaved
	}

	return FuelConsumptionListJSON{
		ConsumptionID:    t.ConsumptionID,
		Status:           t.Status,
		Created_at:       t.Created_at,
		DateFormed:       dateFormed,
		DateCompleted:    dateCompleted,
		CreatorLogin:     creatorLogin,
		ModeratorLogin:   modLogin,
		Origin:           t.Origin,
		Destination:      t.Destination,
		FuelPrice:        t.FuelPrice,
		TotalSaved:       totalSaved,
		FuelEntriesCount: fuelEntriesCount, // ✅ Теперь 0 для не завершённых
	}
}

// ✅ ИСПРАВЛЕНА функция для деталей заявки
func FuelConsumptionToDetailJSON(t ds.FuelConsumption, creatorLogin string, moderatorLogin string, entries []FuelConsumptionEntryViewJSON) FuelConsumptionDetailJSON {
	var dateFormed, dateCompleted *time.Time
	if t.DateFormed.Valid {
		dateFormed = &t.DateFormed.Time
	}
	if t.DateCompleted.Valid {
		dateCompleted = &t.DateCompleted.Time
	}

	var modLogin *string
	if moderatorLogin != "" {
		modLogin = &moderatorLogin
	}

	// ✅ Условие: total_saved = 0, если статус ≠ "завершён"
	totalSaved := float64(0)
	if t.Status == ds.StatusCompleted {
		totalSaved = t.TotalSaved // t.TotalSaved — это float64
	}

	return FuelConsumptionDetailJSON{
		ConsumptionID:  t.ConsumptionID,
		Status:         t.Status,
		Created_at:     t.Created_at,
		DateFormed:     dateFormed,
		DateCompleted:  dateCompleted,
		CreatorLogin:   creatorLogin,
		ModeratorLogin: modLogin,
		Origin:         t.Origin,
		Destination:    t.Destination,
		FuelPrice:      t.FuelPrice,
		TotalSaved:     totalSaved, // ✅ Всегда 0 для не завершённых
		Entries:        entries,
	}
}