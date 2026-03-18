package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"metoda/internal/app/ds"
	"metoda/internal/app/repository"
	"metoda/internal/app/serializer"
	"net/http"
	"strconv"
	"time"
)

// ─── HTML Pages (1 страница: заявка) ────────────────────────────────────────
func (h *Handler) FuelConsumptionPage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/modes")
		return
	}

	t, err := h.Repository.GetFuelConsumptionByID(id)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/modes")
		return
	}

	if t.Status == ds.StatusDeleted {
		ctx.Redirect(http.StatusFound, "/modes")
		return
	}

	if t.Status == ds.StatusDraft && int(t.CreatorID) != repository.GetUserID() {
		ctx.Redirect(http.StatusFound, "/modes")
		return
	}

	items, err := h.Repository.GetFuelConsumptionEntriesAPI(t.ConsumptionID)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/modes")
		return
	}

	ctx.HTML(http.StatusOK, "fuel_consumption.html", gin.H{
		"consumption_id": t.ConsumptionID,
		"modes":          items,
		"minioBase":      minioBaseURL,
		"origin":         t.Origin,
		"destination":    t.Destination,
		"fuel_price":     t.FuelPrice,
		"date_created":   t.DateCreate.Format("02.01.2006 15:04"),
		"status":         t.Status,
		"total_saved":    t.TotalSaved,
	})
}

// ─── API: Fuel Consumptions (7 методов) ─────────────────────────────────────
func (h *Handler) GetFuelConsumptionCart(ctx *gin.Context) {
	creatorID := uint(repository.GetUserID())
	id, count, err := h.Repository.GetCartInfo(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if id == 0 {
		ctx.JSON(http.StatusOK, gin.H{"status": "no_draft", "modes_count": 0})
		return
	}

	ctx.JSON(http.StatusOK, serializer.CartJSON{
		ConsumptionID: id,
		ModesCount:    count,
	})
}

func (h *Handler) GetFuelConsumptions(ctx *gin.Context) {
	var from, to time.Time

	if s := ctx.Query("from_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат from_date"))
			return
		}
		from = t
	}

	if s := ctx.Query("to_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат to_date"))
			return
		}
		to = t
	}

	status := ctx.Query("status")
	list, err := h.Repository.GetAllFuelConsumptions(from, to, status)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	resp := make([]serializer.FuelConsumptionListJSON, 0, len(list))
	for _, t := range list {
		creatorLogin := h.Repository.GetCreatorLogin(t.CreatorID)
		moderatorLogin := h.Repository.GetModeratorLogin(t.ModeratorID)
		entriesCount := h.Repository.GetFuelConsumptionEntriesCount(t.ConsumptionID)
		resp = append(resp, serializer.FuelConsumptionToListJSON(t, creatorLogin, moderatorLogin, entriesCount))
	}
	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) GetFuelConsumption(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный id"))
		return
	}

	t, err := h.Repository.GetFuelConsumptionByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	entries, err := h.Repository.GetFuelConsumptionEntriesAPI(t.ConsumptionID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	creatorLogin := h.Repository.GetCreatorLogin(t.CreatorID)
	moderatorLogin := h.Repository.GetModeratorLogin(t.ModeratorID)

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

	ctx.JSON(http.StatusOK, serializer.FuelConsumptionDetailJSON{
		ConsumptionID:  t.ConsumptionID,
		Status:         t.Status,
		DateCreate:     t.DateCreate,
		DateFormed:     dateFormed,
		DateCompleted:  dateCompleted,
		CreatorLogin:   creatorLogin,
		ModeratorLogin: modLogin,
		Origin:         t.Origin,
		Destination:    t.Destination,
		FuelPrice:      t.FuelPrice,
		TotalSaved:     t.TotalSaved,
		Entries:        entries,
	})
}

func (h *Handler) UpdateFuelConsumption(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный id"))
		return
	}

	var j serializer.FuelConsumptionUpdateJSON
	if err := ctx.ShouldBindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	t, err := h.Repository.UpdateFuelConsumptionFields(id, j)
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

func (h *Handler) FormFuelConsumption(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный id"))
		return
	}

	t, err := h.Repository.FormFuelConsumptionAPI(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		}
		return
	}

	creatorLogin := h.Repository.GetCreatorLogin(t.CreatorID)
	moderatorLogin := h.Repository.GetModeratorLogin(t.ModeratorID)
	entriesCount := h.Repository.GetFuelConsumptionEntriesCount(t.ConsumptionID)

	ctx.JSON(http.StatusOK, serializer.FuelConsumptionToListJSON(t, creatorLogin, moderatorLogin, entriesCount))
}

func (h *Handler) FinishFuelConsumption(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный id"))
		return
	}

	var j serializer.FinishJSON
	if err := ctx.ShouldBindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	t, err := h.Repository.FinishFuelConsumptionAPI(id, j.Status)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusBadRequest, err)
		}
		return
	}

	creatorLogin := h.Repository.GetCreatorLogin(t.CreatorID)
	moderatorLogin := h.Repository.GetModeratorLogin(t.ModeratorID)
	entriesCount := h.Repository.GetFuelConsumptionEntriesCount(t.ConsumptionID)

	ctx.JSON(http.StatusOK, serializer.FuelConsumptionToListJSON(t, creatorLogin, moderatorLogin, entriesCount))
}

// ✅ DELETE /api/fuel-consumptions/:id
func (h *Handler) DeleteFuelConsumption(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный id"))
		return
	}

	// ✅ Вызываем правильную функцию репозитория
	if err := h.Repository.DeleteFuelConsumptionAPI(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
