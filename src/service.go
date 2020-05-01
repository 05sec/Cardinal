package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
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
	s.install()
	s.initMySQL()
	s.initManager()
	s.initStore()
	s.initTimer()
	s.Router = s.initRouter()

	panic(s.Router.Run(conf.Get().Port))
}
