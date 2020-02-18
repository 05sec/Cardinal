package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// DownAction is a gorm model for database table `down_actions`.
type DownAction struct {
	gorm.Model

	TeamID      uint
	ChallengeID uint
	GameBoxID   uint
	Round       int
}

// CheckDown is the gamebox check down handler for bots.
func (s *Service) CheckDown(c *gin.Context) (int, interface{}) {
	// Check down is forbidden if the competition isn't start.
	if s.Timer.Status != "on" {
		return s.makeErrJSON(403, 40300, "比赛未开始")
	}

	type InputForm struct {
		GameBoxID uint `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// Does it check down one gamebox repeatedly in one round?
	var repeatCheck DownAction
	s.Mysql.Model(&DownAction{}).Where(&DownAction{
		GameBoxID: inputForm.GameBoxID,
		Round:     s.Timer.NowRound,
	}).Find(&repeatCheck)
	if repeatCheck.ID != 0 {
		return s.makeErrJSON(403, 40300, "重复 Check down，已忽略")
	}

	// Check the gamebox is existed or not.
	var gameBox GameBox
	s.Mysql.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: inputForm.GameBoxID}}).Find(&gameBox)
	if gameBox.ID == 0 {
		return s.makeErrJSON(403, 40300, "GameBox 不存在！")
	}

	// No problem! Update the gamebox status to down.
	s.Mysql.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: gameBox.ID}}).Update(&GameBox{IsDown: true})

	tx := s.Mysql.Begin()
	if tx.Create(&DownAction{
		TeamID:      gameBox.TeamID,
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
