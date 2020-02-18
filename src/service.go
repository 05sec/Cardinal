package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/patrickmn/go-cache"
)

type Service struct {
	Conf   *Config
	Mysql  *gorm.DB
	Timer  *Timer
	Store  *cache.Cache
	Router *gin.Engine
}

func (s *Service) init() {
	s.install()
	s.initConfig()
	s.initMySQL()
	s.initStore()
	s.initTimer()
	s.initRouter()
}
