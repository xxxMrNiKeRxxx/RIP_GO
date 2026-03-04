package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetFuelConsumption(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := uint(1)

	// Получаем заявку (3 значения: app, items, err)
	app, items, err := h.Repository.GetFuelConsumption(id, creatorID)
	if err != nil {
		logrus.Error(err)
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Расчёт итогов
	totalWithoutCruise := 0.0
	totalSaved := 0.0
	for _, item := range items {
		fuelWithoutCruise := (item.Mode.BaseConsumption / 100) * float64(item.RouteDistance)
		totalWithoutCruise += fuelWithoutCruise
		totalSaved += item.FuelSaved
	}
	totalWithCruise := totalWithoutCruise - totalSaved
	percentSaved := 0.0
	if totalWithoutCruise > 0 {
		percentSaved = (totalSaved / totalWithoutCruise) * 100
	}
	moneySaved := totalSaved * app.FuelPrice

	h.Repository.UpdateTotalSaved(app.ConsumptionID, totalSaved)
	
	ctx.HTML(http.StatusOK, "fuel_consumption.html", gin.H{
		"fuel_consumption":       app,
		"fuel_consumption_modes": items,
		"fuel_consumption_id":    id,
		"routes_count":           len(items),
		"total_without_cruise":   fmt.Sprintf("%.2f", totalWithoutCruise),
		"total_with_cruise":      fmt.Sprintf("%.2f", totalWithCruise),
		"fuel_saved":             fmt.Sprintf("%.2f", totalSaved),
		"percent_saved":          fmt.Sprintf("%.1f", percentSaved),
		"money_saved":            fmt.Sprintf("%.0f", moneySaved),
		"fuel_price":             app.FuelPrice,
		"minioUrl":               h.Config.MinioURL,
		"is_draft":               app.Status == "draft",
	})
}

func (h *Handler) AddToFuelConsumption(ctx *gin.Context) {
	modeIDStr := ctx.PostForm("mode_id")
	modeID, err := strconv.Atoi(modeIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := uint(1)
	routeDistance := 300
	if dist := ctx.PostForm("route_distance"); dist != "" {
		routeDistance, _ = strconv.Atoi(dist)
	}

	err = h.Repository.AddModeToConsumption(uint(modeID), creatorID, routeDistance)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) DeleteFuelConsumption(ctx *gin.Context) {
	appIDStr := ctx.PostForm("fuel_consumption_id")
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Логическое удаление через SQL UPDATE (требование лабы)
	err = h.Repository.DeleteFuelConsumption(uint(appID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Перенаправляем на главную (без заявки)
	ctx.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) UpdateRouteDistance(ctx *gin.Context) {
	appIDStr := ctx.PostForm("fuel_consumption_id")
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	modeIDStr := ctx.PostForm("mode_id")
	modeID, err := strconv.Atoi(modeIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	distanceStr := ctx.PostForm("route_distance")
	routeDistance, err := strconv.Atoi(distanceStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.UpdateRouteDistance(uint(appID), uint(modeID), routeDistance)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/fuel_consumption/"+appIDStr)
}

func (h *Handler) MoveMode(ctx *gin.Context) {
	appIDStr := ctx.PostForm("fuel_consumption_id")
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	modeIDStr := ctx.PostForm("mode_id")
	modeID, err := strconv.Atoi(modeIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	directionStr := ctx.PostForm("direction")
	var direction int
	if directionStr == "up" {
		direction = -1
	} else if directionStr == "down" {
		direction = 1
	} else {
		h.errorHandler(ctx, http.StatusBadRequest, nil)
		return
	}

	err = h.Repository.MoveModeInConsumption(uint(appID), uint(modeID), direction)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/fuel_consumption/"+appIDStr)
}

func (h *Handler) CompleteFuelConsumption(ctx *gin.Context) {
	appIDStr := ctx.PostForm("fuel_consumption_id")
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.CompleteFuelConsumption(uint(appID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/fuel_consumption/"+appIDStr)
}
