package main

import "github.com/jinzhu/gorm"

type Team struct {
	gorm.Model

	Name     string
	Password string
	Score    int64
}
