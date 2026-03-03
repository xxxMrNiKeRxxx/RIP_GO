package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func (h *Handler) GetDrivingModes(ctx *gin.Context) {
	searchText := ctx.Query("search")
	searchField := ctx.DefaultQuery("field", "name")

	var modes []repository.DrivingMode
	var err error
	if searchText == "" {
		modes, err = h.Repository.GetDrivingModes()
	} else {
		modes, err = h.Repository.GetDrivingModesBySearch(searchText, searchField)
	}
	if err != nil {
		logrus.Error(err)
		modes = []repository.DrivingMode{}
	}

	consumptions, _ := h.Repository.GetFuelConsumptions()
	routesCount := 0
	if len(consumptions) > 0 {
		routesCount = consumptions[0].RoutesCount
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"driving_modes": modes,
		"search":        searchText,
		"searchField":   searchField,
		"routes_count":  routesCount,
	})
}

// isNumeric проверяет, является ли строка числом
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func (h *Handler) GetDrivingMode(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Неверный идентификатор")
		return
	}

	mode, err := h.Repository.GetDrivingMode(id)
	if err != nil {
		ctx.String(http.StatusNotFound, "Режим не найден")
		return
	}

	routeInfo, _ := h.Repository.GetFuelConsumptionForMode(id)
	hasRouteInfo := routeInfo != nil

	consumptions, _ := h.Repository.GetFuelConsumptions()
	routesCount := 0
	if len(consumptions) > 0 {
		routesCount = consumptions[0].RoutesCount
	}

	ctx.HTML(http.StatusOK, "mode.html", gin.H{
		"driving_mode":   mode,
		"route_info":     routeInfo,
		"has_route_info": hasRouteInfo,
		"routes_count":   routesCount,
	})
}

func (h *Handler) GetFuelConsumption(ctx *gin.Context) {
	idStr := ctx.Param("id")
	_, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Неверный идентификатор заявки")
		return
	}

	consumption, err := h.Repository.GetFuelConsumption(1)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка загрузки заявки")
		return
	}

	routesCount := consumption.RoutesCount
	totalWithoutCruise := 0.0
	totalSaved := 0.0
	fuelPrice := 55.0

	for _, cm := range consumption.Modes {
		totalWithoutCruise += cm.FuelWithoutCruise
		totalSaved += cm.FuelSaved
	}
	totalWithCruise := totalWithoutCruise - totalSaved
	percentSaved := 0.0
	if totalWithoutCruise > 0 {
		percentSaved = (totalSaved / totalWithoutCruise) * 100
	}
	moneySaved := totalSaved * fuelPrice

	ctx.HTML(http.StatusOK, "fuel_consumption.html", gin.H{
		"fuel_consumption":       consumption,
		"fuel_consumption_modes": consumption.Modes,
		"routes_count":           routesCount,
		"total_without_cruise":   fmt.Sprintf("%.2f", totalWithoutCruise),
		"total_with_cruise":      fmt.Sprintf("%.2f", totalWithCruise),
		"fuel_saved":             fmt.Sprintf("%.2f", totalSaved),
		"percent_saved":          fmt.Sprintf("%.1f", percentSaved),
		"money_saved":            fmt.Sprintf("%.0f", moneySaved),
		"fuel_price":             fuelPrice,
	})
}
