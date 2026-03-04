package handler

import (
	"net/http"
	"strconv"
	"web_backend/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetDrivingModes(ctx *gin.Context) {
	var modes []ds.DrivingMode
	var err error

	searchQuery := ctx.Query("search")
	searchField := ctx.Query("field")

	if searchQuery == "" {
		modes, err = h.Repository.GetDrivingModes()
	} else if searchField == "consumption" {
		consumption, _ := strconv.ParseFloat(searchQuery, 64)
		modes, err = h.Repository.GetDrivingModesByConsumption(consumption)
	} else {
		modes, err = h.Repository.GetDrivingModesBySearch(searchQuery)
	}
	if err != nil {
		logrus.Error(err)
		modes = []ds.DrivingMode{}
	}

	creatorID := uint(1)
	appCount := h.Repository.GetFuelConsumptionCount(creatorID)
	activeAppID := h.Repository.GetActiveFuelConsumptionID(creatorID)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"driving_modes":          modes,
		"search":                 searchQuery,
		"searchField":            searchField,
		"fuel_consumption_count": appCount,
		"fuel_consumption_id":    activeAppID,
		"minioUrl":               h.Config.MinioURL,
		"has_draft":              activeAppID > 0,
	})
}

func (h *Handler) GetDrivingMode(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	mode, err := h.Repository.GetDrivingMode(id)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	creatorID := uint(1)
	activeAppID := h.Repository.GetActiveFuelConsumptionID(creatorID)
	appCount := h.Repository.GetFuelConsumptionCount(creatorID)

	ctx.HTML(http.StatusOK, "mode.html", gin.H{
		"driving_mode":        mode,
		"minioUrl":            h.Config.MinioURL,
		"has_draft":           activeAppID > 0,
		"fuel_consumption_id": activeAppID,
		"routes_count":        appCount,
	})
}
