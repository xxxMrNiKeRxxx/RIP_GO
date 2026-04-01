package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings" // ✅ Добавлен импорт strings

	"github.com/gin-gonic/gin"
	
	"metoda/internal/app/repository"
	"metoda/internal/app/serializer"
)

// GetFuelConsumptionEntries — список м-м в заявке
// @Summary Список м-м в заявке
// @Description Состав конструкций; доступ: создатель или модератор.
// @Tags fuel-consumption-modes
// @Produce json
// @Param id path int true "ID заявки"
// @Security ApiKeyAuth
// @Success 200 {array} serializer.FuelConsumptionEntryViewJSON
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /fuel-consumption-modes/:id [get]
func (h *Handler) GetFuelConsumptionEntries(ctx *gin.Context) {
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

	// ✅ Используем authUserIDUint для получения uid
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// ✅ Проверка прав: владелец или модератор
	if int(t.CreatorID) != int(uid) && !isModeratorFromCtx(ctx) {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("%w: нет доступа к чужой заявке", repository.ErrNotAllowed))
		return
	}

	entries, err := h.Repository.GetFuelConsumptionEntriesAPI(t.ConsumptionID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, entries)
}

// AddToFuelConsumption Добавить режим в заявку (черновик)
// @Summary Добавить режим в заявку
// @Description Добавляет режим в черновик заявки; при первом добавлении создаётся заявка (201 + Location). Поддерживает JSON и form-data.
// @Tags fuel-consumption-modes
// @Produce json
// @Param mode_id path int true "ID режима"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Успешно добавлено"
// @Success 201 {object} map[string]interface{} "Заявка создана"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string "Уже в заявке"
// @Failure 500 {object} map[string]string
// @Router /api/fuel-consumption-modes/add/{mode_id} [post]
func (h *Handler) AddToFuelConsumption(ctx *gin.Context) {
	acceptHeader := ctx.GetHeader("Accept")
	isJSON := strings.Contains(acceptHeader, "application/json") // ✅ Используем strings

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

	// ✅ Используем authUserIDUint для получения uid
	uid, err := authUserIDUint(ctx)
	if err != nil {
		if isJSON {
			h.errorHandler(ctx, http.StatusUnauthorized, err)
			return
		}
		ctx.Redirect(http.StatusSeeOther, "/modes")
		return
	}

	// ✅ Используем правильное имя метода: AddModeToCartAPI
	tp, created, err := h.Repository.AddModeToCartAPI(uint(modeID), uint(uid))
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
			ctx.Header("Location", fmt.Sprintf("/api/fuel-consumptions/%d", tp.ConsumptionID))
			status = http.StatusCreated
		}
		ctx.JSON(status, gin.H{
			"status":           "success",
			"consumption_id": tp.ConsumptionID,
		})
		return
	}

	// HTML-редирект
	redirectTo := ctx.GetHeader("Referer")
	if redirectTo == "" {
		redirectTo = "/modes"
	}
	ctx.Redirect(http.StatusSeeOther, redirectTo)
}

// DeleteModeFromConsumption Убрать режим из заявки
// @Summary Удалить режим из заявки
// @Tags fuel-consumption-modes
// @Produce json
// @Param mode_id path int true "ID режима"
// @Param consumption_id path int true "ID заявки"
// @Security ApiKeyAuth
// @Success 200 {object} serializer.FuelConsumptionListJSON "Обновлённая заявка"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel-consumption-modes/{mode_id}/{consumption_id} [delete]
func (h *Handler) DeleteModeFromConsumption(ctx *gin.Context) {
	modeID, err := strconv.Atoi(ctx.Param("mode_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный mode_id"))
		return
	}
	tpID, err := strconv.Atoi(ctx.Param("consumption_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный consumption_id"))
		return
	}

	// ✅ Используем authUserIDUint для получения uid
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// ✅ Используем правильное имя метода: DeleteModeFromCartAPI
	t, err := h.Repository.DeleteModeFromCartAPI(modeID, tpID, uint(uid))
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
	// ✅ Используем правильное имя метода: GetFuelConsumptionEntriesCount
	entriesCount := h.Repository.GetFuelConsumptionEntriesCount(t.ConsumptionID)
	// ✅ Используем правильное имя функции сериализатора: FuelConsumptionToListJSON
	ctx.JSON(http.StatusOK, serializer.FuelConsumptionToListJSON(t, creatorLogin, moderatorLogin, entriesCount))
}

// UpdateFuelConsumptionMode Обновить параметры записи в заявке
// @Summary Обновить запись в заявке
// @Description Дистанция маршрута для пары режим–заявка.
// @Tags fuel-consumption-modes
// @Accept json
// @Produce json
// @Param mode_id path int true "ID режима"
// @Param consumption_id path int true "ID заявки"
// @Param body body serializer.FuelConsumptionModeUpdateJSON true "Поля для обновления"
// @Security ApiKeyAuth
// @Success 200 {object} serializer.FuelConsumptionModeUpdateJSON
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/fuel-consumption-modes/{mode_id}/{consumption_id} [put]
func (h *Handler) UpdateFuelConsumptionMode(ctx *gin.Context) {
	modeID, err := strconv.Atoi(ctx.Param("mode_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный mode_id"))
		return
	}
	tpID, err := strconv.Atoi(ctx.Param("consumption_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный consumption_id"))
		return
	}

	// ✅ Используем правильный тип: serializer.FuelConsumptionEntryUpdateJSON
	var jIn serializer.FuelConsumptionEntryUpdateJSON
	if err := ctx.ShouldBindJSON(&jIn); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// ✅ Используем authUserIDUint для получения uid
	uid, err := authUserIDUint(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// ✅ Создаём правильный тип для передачи в репозиторий
	j := serializer.FuelConsumptionModeUpdateJSON{
		RouteDistance: jIn.RouteDistance,
		// FuelSaved: jIn.FuelSaved, // <- Скорее всего, этого поля нет в FuelConsumptionModeUpdateJSON
		// SortOrder: jIn.SortOrder, // <- Скорее всего, этого поля нет в FuelConsumptionModeUpdateJSON
	}

	// ✅ Используем правильное имя метода: UpdateModeInCartAPI
	item, err := h.Repository.UpdateModeInCartAPI(modeID, tpID, j, uint(uid))
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

	// ✅ Возвращаем правильный тип: serializer.FuelConsumptionModeUpdateJSON
	ctx.JSON(http.StatusOK, serializer.FuelConsumptionModeUpdateJSON{
		RouteDistance: item.RouteDistance,
		// FuelSaved: item.FuelSaved, // добавьте, если поле есть в ds.FuelConsumptionMode и serializer
		// SortOrder: item.SortOrder, // добавьте, если поле есть в ds.FuelConsumptionMode и serializer
		// добавьте другие поля при необходимости
	})
}