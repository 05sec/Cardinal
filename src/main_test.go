package main

import (
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
	"os"
	"time"
)

var service *Service
var managerToken string
var team []struct {
	Name      string `json:"Name"`
	Password  string `json:"Password"`
	Token     string `json:"token"`
	AccessKey string `json:"access_key"` //submit flag
}

func init() {
	gin.SetMode(gin.ReleaseMode)
	service = new(Service)
	service.Conf = &Config{
		Base: Base{
			Title:          "HCTF",
			BeginTime:      time.Now(),
			RestTime:       nil,
			EndTime:        time.Now().Add(12 * time.Hour),
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
	service.initI18n()
	service.initMySQL()
	service.initStore()
	service.initTimer()

	managerToken = service.generateToken()
	team = make([]struct {
		Name      string `json:"Name"`
		Password  string `json:"Password"`
		Token     string `json:"token"`
		AccessKey string `json:"access_key"`
	}, 0)

	// Test manager account e99:qwe1qwe2qwe3
	service.Mysql.Create(&Manager{
		Name:     "e99",
		Password: service.addSalt("qwe1qwe2qwe3"),
		Token:    managerToken,
	})

	service.Router = service.initRouter()
}
