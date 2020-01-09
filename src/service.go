package main

import "github.com/jinzhu/gorm"

type Service struct {
	Conf  *Config
	Mysql *gorm.DB
}

func (s *Service) init() {
	s.initConfig()
	s.initMySQL()
}
