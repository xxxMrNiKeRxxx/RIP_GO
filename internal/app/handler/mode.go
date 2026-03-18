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
	"strings"
)

// ─── HTML Pages (2 страницы: главная + режим) ────────────────────────────────
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

	cartCount := h.Repository.GetCartCount(uint(repository.GetUserID()))
	consumptionID := h.Repository.GetDraftFuelConsumptionID(uint(repository.GetUserID()))

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"modes":         modes,
		"query":         searchQuery,
		"cartCount":     cartCount,
		"consumptionID": consumptionID,
		"minioBase":     minioBaseURL,
	})
}

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

// ─── API: Modes (3 метода) ──────────────────────────────────────────────────
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
	ctx.JSON(http.StatusCreated, serializer.ModeToJSON(t))
}
