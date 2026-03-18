package serializer

import (
	"metoda/internal/app/ds"
	"time"
)

type FuelConsumptionListJSON struct {
	ConsumptionID    uint       `json:"consumption_id"`
	Status           string     `json:"status"`
	DateCreate       time.Time  `json:"date_create"`
	DateFormed       *time.Time `json:"date_formed"`
	DateCompleted    *time.Time `json:"date_completed"`
	CreatorLogin     string     `json:"creator_login"`
	ModeratorLogin   *string    `json:"moderator_login"`
	Origin           string     `json:"origin"`
	Destination      string     `json:"destination"`
	FuelPrice        float64    `json:"fuel_price"`
	TotalSaved       float64    `json:"total_saved"`
	FuelEntriesCount int        `json:"fuel_entries_count"`
}

type FuelConsumptionDetailJSON struct {
	ConsumptionID  uint                           `json:"consumption_id"`
	Status         string                         `json:"status"`
	DateCreate     time.Time                      `json:"date_create"`
	DateFormed     *time.Time                     `json:"date_formed"`
	DateCompleted  *time.Time                     `json:"date_completed"`
	CreatorLogin   string                         `json:"creator_login"`
	ModeratorLogin *string                        `json:"moderator_login"`
	Origin         string                         `json:"origin"`
	Destination    string                         `json:"destination"`
	FuelPrice      float64                        `json:"fuel_price"`
	TotalSaved     float64                        `json:"total_saved"`
	Entries        []FuelConsumptionEntryViewJSON `json:"entries"`
}

type FuelConsumptionEntryViewJSON struct {
	ID              uint    `json:"id"`
	ModeID          uint    `json:"mode_id"`
	ModeName        string  `json:"mode_name"`
	BaseConsumption float64 `json:"base_consumption"`
	EconomyPercent  float64 `json:"economy_percent"`
	ImageKey        string  `json:"image_key"`
	RouteDistance   int     `json:"route_distance"`
	FuelSaved       float64 `json:"fuel_saved"`
	SortOrder       int     `json:"sort_order"`
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

	return FuelConsumptionListJSON{
		ConsumptionID:    t.ConsumptionID,
		Status:           t.Status,
		DateCreate:       t.DateCreate,
		DateFormed:       dateFormed,
		DateCompleted:    dateCompleted,
		CreatorLogin:     creatorLogin,
		ModeratorLogin:   modLogin,
		Origin:           t.Origin,
		Destination:      t.Destination,
		FuelPrice:        t.FuelPrice,
		TotalSaved:       t.TotalSaved,
		FuelEntriesCount: entriesCount,
	}
}
