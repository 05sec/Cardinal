package main

import "github.com/jinzhu/gorm"

// 攻击记录
type AttackAction struct {
	gorm.Model

	TeamID         uint // 被攻击者
	GameBoxID      uint // 被攻击者靶机
	AttackerTeamID uint // 攻击者
	Round          int
}

// CheckDown 记录
type DownAction struct {
	gorm.Model

	TeamID      uint
	ChallengeID uint
	GameBoxID   uint
	Round       int
}
