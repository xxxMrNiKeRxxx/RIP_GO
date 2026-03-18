package serializer

// FuelConsumptionModeUpdateJSON — для изменения М-М связи (route_distance)
type FuelConsumptionModeUpdateJSON struct {
	RouteDistance int     `json:"route_distance"`
	FuelSaved     float64 `json:"fuel_saved"`
	SortOrder     int     `json:"sort_order"`
}

// FuelConsumptionEntryUpdateJSON — для обновления записи в корзине
type FuelConsumptionEntryUpdateJSON struct {
	RouteDistance int `json:"route_distance"`
}
