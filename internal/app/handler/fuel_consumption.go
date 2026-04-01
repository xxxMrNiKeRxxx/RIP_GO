package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
	"metoda/internal/app/ds"
	"metoda/internal/app/repository"
	"metoda/internal/app/serializer"
)

// ─── HTML Pages ────────────────────────────────────────────────────────────

// FuelConsumptionPage Страница заявки (HTML)
// @Summary Страница заявки на проверку экономии
// @Tags fuel_consumption-html
// @Produce html
// @Param id path int true "ID заявки"
// @Success 200 {string} string "HTML-страница заявки"
// @Failure 302 "Редирект при ошибке или отсутствии доступа"
// @Router /fuel_consumption/{id} [get]
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

	// ✅ Используем authUserIDUint вместо repository.GetUserID()
	uid, err := authUserIDUint(ctx)
	if err != nil {
		// Обработка ошибки авторизации, например, редирект
		ctx.Redirect(http.StatusFound, "/modes")
		return
	}

	if t.Status == ds.StatusDraft && int(t.CreatorID) != int(uid) {
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
		"date_created":   t.Created_at.Format("02.01.2006 15:04"),
		"status":         t.Status,
		"total_saved":    t.TotalSaved,
	})
}

// ─── API: Fuel Consumptions ───────────────────────────────────────────────────

// GetFuelConsumptionCart Корзина (черновик заявки)
// @Summary Корзина (черновик заявки)
// @Description Если черновика нет — возвращает статус "no_draft".
// @Tags fuel_consumptions
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} serializer.CartJSON "CartJSON (tire_pressure_id, tires_count) или no_draft"
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel_consumptions/cart [get]
func (h *Handler) GetFuelConsumptionCart(ctx *gin.Context) {
	// ✅ Используем authUserIDUint вместо repository.GetUserID()
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	creatorID := uint(uid)

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

// GetFuelConsumptions Список заявок
// @Summary Список заявок на проверку экономии
// @Description Фильтры по дате и статусу. Пользователь видит свои; модератор — все.
// @Tags fuel_consumptions
// @Produce json
// @Param created_at query string false "Начало диапазона даты (YYYY-MM-DD)"
// @Param date_completed query string false "Конец диапазона даты (YYYY-MM-DD)"
// @Param status query string false "Фильтр по статусу"
// @Security ApiKeyAuth
// @Success 200 {array} serializer.FuelConsumptionListJSON
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel-consumptions [get]
func (h *Handler) GetFuelConsumptions(ctx *gin.Context) {
	var from, to time.Time
	if s := ctx.Query("created_at"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат created_at"))
			return
		}
		from = t
	}

	if s := ctx.Query("date_completed"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат to_date"))
			return
		}
		to = t
	}
	status := ctx.Query("status")

	// ✅ Используем authUserIDUint и isModeratorFromCtx
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	isMod := isModeratorFromCtx(ctx)

	// ✅ Передаём viewerID и isModerator в репозиторий
	list, err := h.Repository.GetAllFuelConsumptions(from, to, status, uint(uid), isMod)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	resp := make([]serializer.FuelConsumptionListJSON, 0, len(list))
	for _, t := range list { // ✅ Теперь list определён
		creatorLogin := h.Repository.GetCreatorLogin(t.CreatorID)
		moderatorLogin := h.Repository.GetModeratorLogin(t.ModeratorID)
		entriesCount := h.Repository.GetFuelConsumptionEntriesCount(t.ConsumptionID)
		resp = append(resp, serializer.FuelConsumptionToListJSON(t, creatorLogin, moderatorLogin, entriesCount))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetFuelConsumption Детальная заявка
// @Summary Заявка по ID
// @Description Данные по заявке; доступ: создатель или модератор.
// @Tags fuel_consumptions
// @Produce json
// @Param id path int true "ID заявки"
// @Security ApiKeyAuth
// @Success 200 {object} serializer.FuelConsumptionDetailJSON
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel_consumptions/{id} [get]
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

	// ✅ Используем authUserIDUint для проверки прав
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// Проверка прав доступа
	if t.CreatorID != uint(uid) && !isModeratorFromCtx(ctx) {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("%w: нет доступа к чужой заявке", repository.ErrNotAllowed))
		return
	}

	entries, err := h.Repository.GetFuelConsumptionEntriesAPI(t.ConsumptionID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	creatorLogin := h.Repository.GetCreatorLogin(t.CreatorID)
	moderatorLogin := h.Repository.GetModeratorLogin(t.ModeratorID)

	// ❌ УДАЛЕНО: неиспользуемые переменные dateFormed, dateCompleted, modLogin
	ctx.JSON(http.StatusOK, serializer.FuelConsumptionToDetailJSON(t, creatorLogin, moderatorLogin, entries))
}

// UpdateFuelConsumption Обновление полей заявки
// @Summary Обновить заявку
// @Tags fuel_consumptions
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param body body serializer.FuelConsumptionUpdateJSON true "Поля для обновления"
// @Security ApiKeyAuth
// @Success 200 {object} serializer.FuelConsumptionListJSON
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel_consumptions/{id} [put]
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

	// ✅ Получаем userID из контекста
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// ✅ Передаём userID как currentUserID в репозиторий
	t, err := h.Repository.UpdateFuelConsumptionFields(id, j, uint(uid))
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

	// ✅ Вызов с правильным количеством аргументов
	ctx.JSON(http.StatusOK, serializer.FuelConsumptionToListJSON(t, creatorLogin, moderatorLogin, entriesCount))
}

// FormFuelConsumption Оформить заявку (из черновика)
// @Summary Оформить заявку
// @Tags fuel_consumptions
// @Produce json
// @Param id path int true "ID заявки"
// @Security ApiKeyAuth
// @Success 200 {object} serializer.FuelConsumptionListJSON
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel_consumptions/{id}/form [put]
func (h *Handler) FormFuelConsumption(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный id"))
		return
	}

	// ✅ Получаем userID из контекста
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// ✅ Передаём userID как currentUserID в репозиторий
	t, err := h.Repository.FormFuelConsumptionAPI(id, uint(uid))
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

// FinishFuelConsumption Завершить заявку (модератор)
// @Summary Завершить заявку
// @Description Установка итогового статуса; только для модератора.
// @Tags fuel_consumptions
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param body body serializer.FinishJSON true "Статус завершения"
// @Security ApiKeyAuth
// @Success 200 {object} serializer.FuelConsumptionListJSON
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel-consumptions/{id}/finish [put]
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

	// ✅ Получаем userID из контекста
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// ✅ Передаём userID как currentUserID в репозиторий
	t, err := h.Repository.FinishFuelConsumptionAPI(id, j.Status, uint(uid))
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

// DeleteFuelConsumption Удалить заявку
// @Summary Удалить заявку
// @Tags fuel_consumptions
// @Produce json
// @Param id path int true "ID заявки"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "status: deleted"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel_consumptions/{id} [delete]
func (h *Handler) DeleteFuelConsumption(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный id"))
		return
	}

	// ✅ Получаем userID из контекста
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// ✅ Передаём userID как currentUserID в репозиторий
	if err := h.Repository.DeleteFuelConsumptionAPI(id, uint(uid)); err != nil {
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