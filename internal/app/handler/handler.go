package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"metoda/internal/app/repository"
	"net/http"
	"os"
	"strconv"
)

// ✅ ИСПРАВЛЕНО: /tires → /services
const minioBaseURL = "http://localhost:9000/services"

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) getMinioURL() string {
	if url := os.Getenv("MINIO_URL"); url != "" {
		return url
	}
	return minioBaseURL
}

// ─── Register Routes ─────────────────────────────────────────────────────────
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// ✅ Добавить редирект с "/" на "/modes"

	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/modes")
	})

	// 3 HTML страницы
	router.GET("/modes", h.Index)
	router.GET("/mode/:id", h.ModePage)
	router.GET("/fuel-consumption/:id", h.FuelConsumptionPage)

	// Driving Modes (3 метода)
	router.GET("/api/modes", h.GetModes)
	router.GET("/api/modes/:id", h.GetMode)
	router.POST("/api/modes", h.CreateMode)

	// Fuel Consumptions (7 методов)
	router.GET("/api/fuel-consumptions/cart", h.GetFuelConsumptionCart)
	router.GET("/api/fuel-consumptions", h.GetFuelConsumptions)
	router.GET("/api/fuel-consumptions/:id", h.GetFuelConsumption)
	router.PUT("/api/fuel-consumptions/:id", h.UpdateFuelConsumption)
	router.PUT("/api/fuel-consumptions/:id/form", h.FormFuelConsumption)
	router.PUT("/api/fuel-consumptions/:id/finish", h.FinishFuelConsumption)
	router.DELETE("/api/fuel-consumptions/:id", h.DeleteFuelConsumption)

	// Fuel Consumption Modes (3 метода) - ✅ ПРАВИЛЬНЫЙ МАРШРУТ
	router.POST("/api/fuel-consumption-modes/add/:mode_id", h.AddToFuelConsumption)
	router.DELETE("/api/fuel-consumption-modes/:mode_id/:consumption_id", h.DeleteModeFromConsumption)
	router.PUT("/api/fuel-consumption-modes/:mode_id/:consumption_id", h.UpdateFuelConsumptionMode)

	// Users (3 метода)
	router.POST("/api/users/signup", h.APISignUp)
	router.POST("/api/users/signin", h.APISignIn)
	router.POST("/api/users/signout", h.APISignOut)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	var errorMessage string
	switch {
	case errors.Is(err, repository.ErrNotFound):
		errorMessage = "Не найдено"
	case errors.Is(err, repository.ErrAlreadyExists):
		errorMessage = "Уже существует"
	case errors.Is(err, repository.ErrNotAllowed):
		errorMessage = "Доступ запрещён"
	case errors.Is(err, repository.ErrNoDraft):
		errorMessage = "Черновик не найден"
	default:
		errorMessage = err.Error()
	}
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": errorMessage,
	})
}

// UploadImageToMode — заглушка для демонстрации (без MinIO)
// UploadImageToMode — заглушка для демонстрации (без MinIO)
func (h *Handler) UploadImageToMode(ctx *gin.Context) {
	modeIDStr := ctx.Param("id")
	modeID, err := strconv.Atoi(modeIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid id"))
		return
	}

	// Просто возвращаем success, без загрузки в MinIO
	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Image upload simulated",
		"image_key": fmt.Sprintf("mode_%d_simulated.jpg", modeID),
		"mode_id":   modeID,
	})
}
