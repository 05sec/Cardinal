package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// CheckDown 记录
type DownAction struct {
	gorm.Model

	TeamID      uint
	ChallengeID uint
	GameBoxID   uint
	Round       int
}

func (s *Service) CheckDown(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		TeamID    uint `binding:"required"`
		GameBoxID uint `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// 是否重复 Check
	var repeatCheck DownAction
	s.Mysql.Model(&DownAction{}).Where(&DownAction{
		TeamID:    inputForm.TeamID,
		GameBoxID: inputForm.GameBoxID,
		Round:     s.Timer.NowRound,
	}).Find(&repeatCheck)
	if repeatCheck.ID != 0 {
		return s.makeErrJSON(403, 40300, "重复 Check down，已忽略")
	}

	// 确认 GameBox 信息
	var gameBox GameBox
	s.Mysql.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: inputForm.GameBoxID}, TeamID: inputForm.TeamID}).Find(&gameBox)
	if gameBox.ID == 0 {
		return s.makeErrJSON(403, 40300, "GameBox 不存在！")
	}

	tx := s.Mysql.Begin()
	if tx.Create(&DownAction{
		TeamID:      inputForm.TeamID,
		ChallengeID: gameBox.ChallengeID,
		GameBoxID:   inputForm.GameBoxID,
		Round:       s.Timer.NowRound,
	}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "Server error")
	}
	tx.Commit()
	return s.makeSuccessJSON("success")
}
