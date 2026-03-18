package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"metoda/internal/app/config"
	"metoda/internal/app/dsn"
	"metoda/internal/app/handler"
	"metoda/internal/app/repository"
	"metoda/internal/pkg"
)

func main() {
	// ✅ КРИТИЧНО: Загрузить .env файл ПЕРЕД использованием dsn.FromEnv()
	err := godotenv.Load()
	if err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println("DSN:", postgresString)

	if postgresString == "" {
		logrus.Fatal("DSN is empty! Check .env file")
	}

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)
	app := pkg.NewApp(conf, router, hand)
	app.RunApp()
}
