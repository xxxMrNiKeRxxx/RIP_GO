package main

import (
	"fmt"
	"html/template"
	"strconv"

	"web_backend/internal/app/config"
	"web_backend/internal/app/dsn"
	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"
	"web_backend/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	router.SetFuncMap(template.FuncMap{
		"printf": fmt.Sprintf,
		"str":    func(v interface{}) string { return fmt.Sprint(v) },
		"sub":   func(a, b int) int { return a - b },
		"isMain": func(mainID interface{}, deptID uint) bool {
			if mainID == nil {
				return false
			}
			if p, ok := mainID.(*uint); ok && p != nil {
				return *p == deptID
			}
			return false
		},
		"num": func(v interface{}) float64 {
			if v == nil {
				return 0
			}
			switch x := v.(type) {
			case float64:
				return x
			case *float64:
				if x == nil {
					return 0
				}
				return *x
			default:
				f, _ := strconv.ParseFloat(fmt.Sprint(v), 64)
				return f
			}
		},
	})

	postgresString := dsn.FromEnv()
	logrus.Info("DSN: ", postgresString)

	rep, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	hand := handler.NewHandler(rep, conf)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
