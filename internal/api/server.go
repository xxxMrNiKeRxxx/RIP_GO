package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("Ошибка инициализации репозитория")
	}

	h := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", h.GetDrivingModes)
	r.GET("/mode/:id", h.GetDrivingMode)
	r.GET("/fuel_consumption/:id", h.GetFuelConsumption)

	r.Run()
	log.Println("Server down")
}
