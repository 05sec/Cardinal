package main

import (
	"github.com/jinzhu/gorm"
	"time"
)

type BaseConfig struct {
	gorm.Model `json:"-"`

	Title     string
	BeginTime time.Time
	EndTime   time.Time
	Duration  uint
}
