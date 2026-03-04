package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/config"
	"web_backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
}

func NewHandler(r *repository.Repository, cfg *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Config:     cfg,
	}
}

// RegisterHandler - регистрация всех роутов (ОДИН РАЗ)
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.GetDrivingModes)
	router.GET("/mode/:id", h.GetDrivingMode)
	router.GET("/fuel_consumption/:id", h.GetFuelConsumption)
	router.POST("/fuel_consumption/add", h.AddToFuelConsumption)
	router.POST("/fuel_consumption/update_distance", h.UpdateRouteDistance)
	router.POST("/fuel_consumption/move", h.MoveMode)
	router.POST("/fuel_consumption/delete", h.DeleteFuelConsumption)
	router.POST("/fuel_consumption/complete", h.CompleteFuelConsumption)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
