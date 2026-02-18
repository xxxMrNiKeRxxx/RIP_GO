package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"web_backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// OrderItemView - вспомогательная структура для передачи предвычисленных значений в шаблон
type OrderItemView struct {
	repository.Device
	RouteLength       int
	FuelBase          float64
	FuelBaseFormatted string
	FuelWithCruise    float64
	SavingPercent     float64
}

// GetDevices - GET / - страница списка услуг с фильтрацией
func (h *Handler) GetDevices(ctx *gin.Context) {
	var devices []repository.Device
	var err error

	searchQuery := ctx.Query("query")
	searchField := ctx.DefaultQuery("field", "name")

	if searchQuery == "" {
		devices, err = h.Repository.GetDevices()
		if err != nil {
			logrus.Error("GetDevices error:", err)
			devices = []repository.Device{}
		}
	} else {
		devices, err = h.Repository.FilterDevices(searchQuery, searchField)
		if err != nil {
			logrus.Error("FilterDevices error:", err)
			devices = []repository.Device{}
		}
	}

	cartDevices, err := h.Repository.GetCart()
	cartCount := 0
	if err == nil {
		cartCount = len(cartDevices)
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"devices":     devices,
		"query":       searchQuery,
		"searchField": searchField,
		"cartCount":   cartCount,
		"orderID":     "ORD-CC-001",
	})
}

// GetDevice - GET /order/:id - страница детали услуги
func (h *Handler) GetDevice(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Invalid ID:", err)
		ctx.String(http.StatusBadRequest, "Неверный ID")
		return
	}

	device, err := h.Repository.GetDevice(id)
	if err != nil {
		logrus.Error("Device not found:", err)
		ctx.String(http.StatusNotFound, "Услуга не найдена")
		return
	}

	cartDevices, _ := h.Repository.GetCart()
	cartCount := len(cartDevices)

	ctx.HTML(http.StatusOK, "order.html", gin.H{
		"device":      device,
		"cartCount":   cartCount,
		"orderID":     "ORD-CC-001",
		"query":       ctx.Query("query"),
		"searchField": ctx.Query("field"),
	})
}

// GetCart - GET /cart - страница заявки с расчетом экономии топлива
func (h *Handler) GetCart(ctx *gin.Context) {
	devices, err := h.Repository.GetCart()
	if err != nil {
		logrus.Error("GetCart error:", err)
		devices = []repository.Device{}
	}

	// === РАСЧЕТ ЭКОНОМИИ ТОПЛИВА В КОНТРОЛЛЕРЕ ===
	totalWithoutCruise := 0.0
	totalWithCruise := 0.0
	fuelPrice := 55.0 // цена топлива руб/л

	var itemsView []OrderItemView
	for _, item := range devices {
		// Длина маршрута = ID * 100 км (эмуляция)
		routeLength := item.ID * 100

		// Расход без круиз-контроля: (расход л/100км) * (км) / 100
		fuelBase := (item.Power / 100) * float64(routeLength)

		// Процент экономии в зависимости от режима
		savingPercent := getSavingPercent(item.ServiceType)

		// Расход с круиз-контролем
		fuelWithCruise := fuelBase * (1 - savingPercent/100)

		// Суммируем для итогов
		totalWithoutCruise += fuelBase
		totalWithCruise += fuelWithCruise

		// Добавляем в список для отображения
		itemsView = append(itemsView, OrderItemView{
			Device:            item,
			RouteLength:       routeLength,
			FuelBase:          fuelBase,
			FuelBaseFormatted: fmt.Sprintf("%.2f", fuelBase),
			FuelWithCruise:    fuelWithCruise,
			SavingPercent:     savingPercent,
		})
	}

	// Итоговые значения
	fuelSaved := totalWithoutCruise - totalWithCruise
	percentSaved := 0.0
	if totalWithoutCruise > 0 {
		percentSaved = (fuelSaved / totalWithoutCruise) * 100
	}
	moneySaved := fuelSaved * fuelPrice

	// Готовое значение результата вычислений (для поля в заявке)
	calculationResult := fmt.Sprintf("Экономия: %.2f л (%.1f%%) | %.0f ₽",
		fuelSaved, percentSaved, moneySaved)

	ctx.HTML(http.StatusOK, "cart.html", gin.H{
		"service_devices":    itemsView,
		"cartCount":          len(itemsView),
		"orderID":            "ORD-CC-001",
		"calculationResult":  calculationResult,
		"totalWithoutCruise": fmt.Sprintf("%.2f", totalWithoutCruise),
		"totalWithCruise":    fmt.Sprintf("%.2f", totalWithCruise),
		"fuelSaved":          fmt.Sprintf("%.2f", fuelSaved),
		"percentSaved":       fmt.Sprintf("%.1f", percentSaved),
		"moneySaved":         fmt.Sprintf("%.0f", moneySaved),
		"fuelPrice":          fuelPrice,
		"query":              ctx.Query("query"),
		"searchField":        ctx.Query("field"),
	})
}

// getSavingPercent - определяет процент экономии по типу режима
func getSavingPercent(serviceType string) float64 {
	switch strings.ToLower(serviceType) {
	case "city":
		return 5.0 // город: экономия ~5%
	case "highway":
		return 15.0 // трасса: экономия ~15%
	default:
		return 10.0 // смешанный: экономия ~10%
	}
}
