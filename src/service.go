package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
	"github.com/vidar-team/Cardinal/src/asteroid"
	"github.com/vidar-team/Cardinal/src/conf"
)

// Service is the main struct contains all the part of the Cardinal.
type Service struct {
	Mysql  *gorm.DB
	Timer  *Timer
	Store  *cache.Cache
	Router *gin.Engine
}

func (s *Service) init() {
	// Check install or not.
	s.install()

	// Load config file.
	conf.Init()

	// Init database.
	s.initMySQL()

	// Refresh the dynamic config from the database.
	s.RefreshConfig()

	// Unity3D Asteroid
	asteroid.InitAsteroid(s.asteroidGreetData)

	// Check manager account exist or not.
	s.initManager()

	// Cache
	s.initStore()

	// Game timer.
	s.initTimer()

	// Web router.
	s.Router = s.initRouter()

	panic(s.Router.Run(conf.Get().Port))
}
