package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"os"
	"time"
)

var service *Service
var managerToken string

func init() {
	fmt.Println("BEGIN")
	gin.SetMode(gin.ReleaseMode)
	service = new(Service)
	service.Conf = &Config{
		Base: Base{
			Title:          "HCTF",
			BeginTime:      time.Now(),
			RestTime:       nil,
			EndTime:        time.Now(),
			Duration:       10,
			Port:           ":19999",
			Salt:           randstr.String(64),
			FlagPrefix:     "hctf{",
			FlagSuffix:     "}",
			CheckDownScore: 10,
			AttackScore:    10,
		},
		MySQL: MySQL{
			DBHost:     "127.0.0.1:3306",
			DBUsername: "root",
			DBPassword: os.Getenv("TEST_DB_PASSWORD"),
			DBName:     os.Getenv("TEST_DB_NAME"),
		},
	}
	service.initMySQL()
	service.initStore()

	managerToken = service.generateToken()
	// Test manager account e99:qwe1qwe2qwe3
	service.Mysql.Create(&Manager{
		Name:     "e99",
		Password: service.addSalt("qwe1qwe2qwe3"),
		Token:    managerToken,
	})

	service.Router = service.initRouter()
}
