package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"metoda/internal/app/ds"
	"metoda/internal/app/repository"
	"metoda/internal/app/serializer"
)

// ─── HTML Pages ────────────────────────────────────────────────────────────

// Index Главная страница со типов движения
// @Summary Главная страница (типы движения)
// @Tags mode-html
// @Produce html
// @Param query query string false "Поиск по названию режима"
// @Success 200 {string} string "HTML-страница"
// @Router /modes [get]
func (h *Handler) Index(ctx *gin.Context) {
	var modes []ds.DrivingMode
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		modes, err = h.Repository.GetAllModes()
	} else {
		modes, err = h.Repository.SearchModesByName(searchQuery)
	}
	if err != nil {
		modes = []ds.DrivingMode{}
	}

	uid, err := authUserIDUint(ctx)
	if err != nil {
		// Обработка ошибки, если не авторизован
		// logrus.Error(err)
		// ctx.AbortWithStatus(http.StatusUnauthorized)
		// return
		// Или установить uid = 0 и показать пустую корзину
		uid = 0
	}

	cartCount := h.Repository.GetCartCount(uint(uid))
	consumptionID := h.Repository.GetDraftFuelConsumptionID(uint(uid))

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"modes":         modes,
		"query":         searchQuery,
		"cartCount":     cartCount,
		"consumptionID": consumptionID,
		"minioBase":     minioBaseURL,
	})
}

// ModePage Страница одного режима
// @Summary Страница режима по ID
// @Tags modes-html
// @Produce html
// @Param id path int true "ID режима"
// @Success 200 {string} string "HTML-страница"
// @Failure 404 {object} map[string]string
// @Router /mode/{id} [get]
func (h *Handler) ModePage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	mode, err := h.Repository.GetModeByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "mode.html", gin.H{
		"mode":      mode,
		"minioBase": minioBaseURL,
	})
}

// ─── API: Modes ────────────────────────────────────────────────────────────

// GetModes Список режимов (API)
// @Summary Список режимов
// @Tags Modes
// @Produce json
// @Param query query string false "Поиск по названию режима"
// @Success 200 {array} serializer.ModeJSON
// @Failure 500 {object} map[string]string
// @Router /api/modes [get]
func (h *Handler) GetModes(ctx *gin.Context) {
	query := ctx.Query("query")
	var modes []ds.DrivingMode
	var err error

	if query == "" {
		modes, err = h.Repository.GetAllModes()
	} else {
		modes, err = h.Repository.SearchModesByName(query)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	resp := make([]serializer.ModeJSON, 0, len(modes))
	for _, t := range modes {
		resp = append(resp, serializer.ModeToJSON(t))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetMode Один режим по ID (API)
// @Summary режим по ID
// @Tags Modes
// @Produce json
// @Param id path int true "ID режима"
// @Success 200 {object} serializer.ModeJSON
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/modes/{id} [get]
func (h *Handler) GetMode(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный id"))
		return
	}

	t, err := h.Repository.GetModeByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, serializer.ModeToJSON(*t))
}

// CreateMode Создание режима (только модератор)
// @Summary Создать режим
// @Description multipart/form-data или application/json; только для модераторов
// @Tags Modes
// @Accept mpfd
// @Produce json
// @Param Mode_title formData string false "Название режим"
// @Param Mode_material_coefficient formData number false "Коэффициент материала"
// @Param Mode_thickness_coefficient formData number false "Коэффициент толщины"
// @Param description formData string false "Описание"
// @Param photo formData file false "Фото режима"
// @Param video formData file false "Видео режима"
// @Success 201 {object} serializer.ModeJSON
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string "Требуется роль модератора"
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /api/Modes [post]
func (h *Handler) CreateMode(ctx *gin.Context) {
		contentType := ctx.GetHeader("Content-Type")
	var j serializer.ModeJSON

	if strings.HasPrefix(contentType, "application/json") {
		if err := ctx.BindJSON(&j); err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
	} else {
		baseConsumption, _ := strconv.ParseFloat(ctx.PostForm("base_consumption"), 64)
		economyPercent, _ := strconv.ParseFloat(ctx.PostForm("economy_percent"), 64)
		j = serializer.ModeJSON{
			ModeName:        ctx.PostForm("mode_name"),
			BaseConsumption: baseConsumption,
			EconomyPercent:  economyPercent,
			Description:     ctx.PostForm("description"),
			DrivingType:     ctx.PostForm("driving_type"),
		}
	}

	t, err := h.Repository.CreateMode(j)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if photoFile, err := ctx.FormFile("image"); err == nil {
		t, _ = h.Repository.UploadModeImage(ctx, int(t.ModeID), photoFile)
	}
	if videoFile, err := ctx.FormFile("video"); err == nil {
		t, _ = h.Repository.UploadModeVideo(ctx, int(t.ModeID), videoFile)
	}

	ctx.Header("Location", fmt.Sprintf("/api/modes/%d", t.ModeID))
	ctx.JSON(http.StatusCreated, serializer.ModeToJSON(t)) // ✅ t - указатель
}