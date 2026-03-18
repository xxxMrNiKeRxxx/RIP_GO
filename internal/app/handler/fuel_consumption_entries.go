package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"metoda/internal/app/repository"
	"metoda/internal/app/serializer"
	"net/http"
	"strconv"
	"strings"
)

// ─── API: Fuel Consumption Modes (3 метода) ──────────────────────────────────
func (h *Handler) AddToFuelConsumption(ctx *gin.Context) {
	acceptHeader := ctx.GetHeader("Accept")
	isJSON := strings.Contains(acceptHeader, "application/json")

	modeIDStr := ctx.Param("mode_id")
	if modeIDStr == "" {
		modeIDStr = ctx.PostForm("mode_id")
	}
	if modeIDStr == "" {
		if isJSON {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("mode_id is required"))
			return
		}
		ctx.Redirect(http.StatusSeeOther, "/modes")
		return
	}

	modeID, err := strconv.Atoi(modeIDStr)
	if err != nil {
		if isJSON {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		ctx.Redirect(http.StatusSeeOther, "/modes")
		return
	}

	creatorID := uint(repository.GetUserID())
	fc, created, err := h.Repository.AddModeToCartAPI(uint(modeID), creatorID)
	if err != nil {
		if isJSON {
			if errors.Is(err, repository.ErrNotFound) {
				h.errorHandler(ctx, http.StatusNotFound, err)
			} else if errors.Is(err, repository.ErrAlreadyExists) {
				h.errorHandler(ctx, http.StatusConflict, err)
			} else {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
			}
			return
		}
		ctx.Redirect(http.StatusSeeOther, "/modes")
		return
	}

	if isJSON {
		status := http.StatusOK
		if created {
			ctx.Header("Location", fmt.Sprintf("/api/fuel-consumptions/%d", fc.ConsumptionID))
			status = http.StatusCreated
		}
		ctx.JSON(status, gin.H{"status": "success", "consumption_id": fc.ConsumptionID})
		return
	}

	redirectTo := ctx.GetHeader("Referer")
	if redirectTo == "" {
		redirectTo = "/modes"
	}
	ctx.Redirect(http.StatusSeeOther, redirectTo)
}

func (h *Handler) UpdateFuelConsumptionMode(ctx *gin.Context) {
	modeID, err := strconv.Atoi(ctx.Param("mode_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	fcID, err := strconv.Atoi(ctx.Param("consumption_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var j serializer.FuelConsumptionModeUpdateJSON
	if err := ctx.ShouldBindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	item, err := h.Repository.UpdateModeInCartAPI(modeID, fcID, j)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"route_distance": item.RouteDistance,
		"fuel_saved":     item.FuelSaved,
	})
}

func (h *Handler) DeleteModeFromConsumption(ctx *gin.Context) {
	modeID, err := strconv.Atoi(ctx.Param("mode_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	fcID, err := strconv.Atoi(ctx.Param("consumption_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	t, err := h.Repository.DeleteModeFromCartAPI(modeID, fcID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	creatorLogin := h.Repository.GetCreatorLogin(t.CreatorID)
	moderatorLogin := h.Repository.GetModeratorLogin(t.ModeratorID)
	entriesCount := h.Repository.GetFuelConsumptionEntriesCount(t.ConsumptionID)

	ctx.JSON(http.StatusOK, serializer.FuelConsumptionToListJSON(t, creatorLogin, moderatorLogin, entriesCount))
}
