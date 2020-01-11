package main

import "github.com/jinzhu/gorm"

// 攻击记录
type AttackAction struct {
	gorm.Model

	TeamID         uint8 // 被攻击者
	GameBoxID      uint8 // 被攻击者靶机
	AttackerTeamID uint8 // 攻击者
	Round          int
}

// CheckDown 记录
type DownAction struct {
	gorm.Model

	TeamID      uint8
	ChallengeID uint8
	GameBoxID   uint8
	Round       int
}
