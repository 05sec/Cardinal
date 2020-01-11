package main

import "github.com/jinzhu/gorm"

type Score struct {
	gorm.Model

	TeamID    uint8
	GameBoxID uint8
	Round     int
	Reason    string
	Score     uint8		`gorm:"index"`
}
