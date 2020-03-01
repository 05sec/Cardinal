package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
	"github.com/qor/i18n"
)

// Service is the main struct contains all the part of the Cardinal.
type Service struct {
	Conf   *Config
	Mysql  *gorm.DB
	Timer  *Timer
	I18n   *i18n.I18n
	Store  *cache.Cache
	Router *gin.Engine
}

func (s *Service) init() {
	s.initI18n()
	s.install()
	s.initConfig()
	s.initMySQL()
	s.initManager()
	s.initStore()
	s.initTimer()
	s.Router = s.initRouter()

	panic(s.Router.Run(s.Conf.Base.Port))
}
