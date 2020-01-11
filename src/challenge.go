package main

import "github.com/jinzhu/gorm"

// 题目
type Challenge struct {
	gorm.Model
	Title string
}

// 队伍靶机
type GameBox struct {
	gorm.Model
	ChallengeID uint8
	TeamID      uint8

	Description string
	Visible     bool
	Score       int64 // 分数可负
	IsDown      bool
	IsAttacked  bool
}
