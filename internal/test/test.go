package test

import (
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/internal/asteroid"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/dynamic_config"
	"github.com/vidar-team/Cardinal/internal/game"
	"github.com/vidar-team/Cardinal/internal/livelog"
	"github.com/vidar-team/Cardinal/internal/misc/webhook"
	"github.com/vidar-team/Cardinal/internal/route"
	"github.com/vidar-team/Cardinal/internal/store"
	"github.com/vidar-team/Cardinal/internal/timer"
	"github.com/vidar-team/Cardinal/internal/utils"
)

var ManagerToken string
var CheckToken string
var Team []struct {
	Name      string `json:"Name"`
	Password  string `json:"Password"`
	Token     string `json:"token"`
	AccessKey string `json:"access_key"` //submit flag
}
var Router *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)

	// Init MySQL database.
	db.InitMySQL()

	// Refresh the dynamic config from the database.
	dynamic_config.Init()

	// Game timer.
	timer.Init()

	asteroid.Init(game.AsteroidGreetData)

	// Cache
	store.Init()
	webhook.RefreshWebHookStore()

	// Live log
	livelog.Init()

	// Web router.
	Router = route.Init()

	ManagerToken = utils.GenerateToken()
	Team = make([]struct {
		Name      string `json:"Name"`
		Password  string `json:"Password"`
		Token     string `json:"token"`
		AccessKey string `json:"access_key"`
	}, 0)

	// Test manager account e99:qwe1qwe2qwe3
	db.MySQL.Create(&db.Manager{
		Name:     "e99",
		Password: utils.AddSalt("qwe1qwe2qwe3"),
		Token:    ManagerToken,
		IsCheck:  false,
	})
}
