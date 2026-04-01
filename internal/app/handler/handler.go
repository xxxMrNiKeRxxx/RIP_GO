package handler

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"metoda/internal/app/repository"
)

const minioBaseURL = "http://localhost:9090/services"

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
	// ─── HTML страницы (публичные) ──────────────────────────────────────
	router.GET("/modes", h.Index)
	router.GET("/mode/:id", h.ModePage)
	router.GET("/fuel-consumption/:id", h.FuelConsumptionPage)

	// ─── API группа ─────────────────────────────────────────────────────
	api := router.Group("/api")

	// === ПУБЛИЧНЫЕ эндпоинты (без авторизации) ===
	// Чтение каталога + регистрация и вход
	api.GET("/modes", h.GetModes)
	api.GET("/modes/:id", h.GetMode)
	api.POST("/users/signup", h.APISignUp)
	api.POST("/users/signin", h.APISignIn)

	// === ЗАЩИЩЁННЫЕ эндпоинты (требуется авторизация) ===
	needAuth := api.Group("")
	needAuth.Use(h.AuthMiddleware())

	// Выход из системы
	needAuth.POST("/users/signout", h.APISignOut)

	// Fuel Consumptions (черновики/заявки)
	needAuth.GET("/fuel-consumptions/cart", h.GetFuelConsumptionCart)
	needAuth.GET("/fuel-consumptions", h.GetFuelConsumptions)
	needAuth.GET("/fuel-consumptions/:id", h.GetFuelConsumption)
	needAuth.PUT("/fuel-consumptions/:id", h.UpdateFuelConsumption)
	needAuth.PUT("/fuel-consumptions/:id/form", h.FormFuelConsumption)
	needAuth.DELETE("/fuel-consumptions/:id", h.DeleteFuelConsumption)

	// Fuel Consumption Modes  (записи в заявке)
	// ✅ Используем правильные имена методов handler
	needAuth.POST("/fuel-consumption-modes/add/:mode_id", h.AddToFuelConsumption)
	needAuth.DELETE("/fuel-consumption-modes/:mode_id/:consumption_id", h.DeleteModeFromConsumption)
	needAuth.PUT("/fuel-consumption-modes/:mode_id/:consumption_id", h.UpdateFuelConsumptionMode)

	// === ЭНДПОИНТЫ ДЛЯ МОДЕРАТОРОВ ===
	mod := api.Group("")
	mod.Use(h.AuthMiddleware())
	mod.Use(h.RequireModerator())

	// Создание шины (только модератор)
	mod.POST("/modes", h.CreateMode)

	// Завершение заявки (только модератор)
	mod.PUT("/fuel-consumptions/:id/finish", h.FinishFuelConsumption)

	// Swagger
	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	router.Any("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

// ─── Error Handler (расширенный) ─────────────────────────────────────────

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
	case errors.Is(err, repository.ErrUnauthorized):
		errorMessage = "Необходимо войти в систему"
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "description": errorMessage})
		return
	case errors.Is(err, repository.ErrForbidden):
		errorMessage = "Недостаточно прав"
		ctx.JSON(http.StatusForbidden, gin.H{"status": "error", "description": errorMessage})
		return
	case errors.Is(err, repository.ErrInvalidToken):
		errorMessage = "Неверный токен"
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "description": errorMessage})
		return
	case errors.Is(err, repository.ErrTokenExpired):
		errorMessage = "Срок токена истёк"
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "description": errorMessage})
		return
	default:
		errorMessage = err.Error()
	}

	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": errorMessage,
	})
}