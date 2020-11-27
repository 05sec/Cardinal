package bootstrap

import (
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/asteroid"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/dynamic_config"
	"github.com/vidar-team/Cardinal/internal/game"
	"github.com/vidar-team/Cardinal/internal/install"
	"github.com/vidar-team/Cardinal/internal/livelog"
	"github.com/vidar-team/Cardinal/internal/misc"
	"github.com/vidar-team/Cardinal/internal/misc/webhook"
	"github.com/vidar-team/Cardinal/internal/route"
	"github.com/vidar-team/Cardinal/internal/store"
	"github.com/vidar-team/Cardinal/internal/timer"
)

func init() {
	// Init log
	_ = log.NewConsole(100)
}

// LinkStart starts the Cardinal.
func LinkStart() {
	// Install
	install.Init()

	// Config
	conf.Init()

	// Check version
	misc.CheckVersion()

	// Sentry
	misc.Sentry()

	// Init MySQL database.
	db.InitMySQL()

	// Check manager
	install.InitManager()

	// Refresh the dynamic config from the database.
	dynamic_config.Init()

	// Check if the database need update.
	misc.CheckDatabaseVersion()

	// Game timer.
	GameToTimerBridge()
	timer.Init()

	// Cache
	store.Init()
	webhook.RefreshWebHookStore()

	// Unity3D Asteroid
	asteroid.Init(game.AsteroidGreetData)

	// Live log
	livelog.Init()

	// Web router.
	router := route.Init()

	log.Fatal("Failed to start web server: %v", router.Run(conf.Get().Port))
}
