package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Service struct {
	Conf   *Config
	Mysql  *gorm.DB
	Timer  *Timer
	Router *gin.Engine
}

func (s *Service) init() {
	s.initConfig()
	s.initMySQL()
	s.initTimer()
	s.initRouter()

}
