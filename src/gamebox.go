package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// 队伍靶机
type GameBox struct {
	gorm.Model
	ChallengeID uint
	TeamID      uint

	Description string
	Visible     bool
	Score       int64 // 分数可负
	IsDown      bool
	IsAttacked  bool
}

func (s *Service) NewGameBoxes(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ChallengeID uint   `binding:"required"`
		TeamID      uint   `binding:"required"`
		Description string `binding:"required"`
	}
	var inputForm []InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	for _, item := range inputForm {
		var count int

		// 检查 ChallengeID
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: item.ChallengeID}}).Count(&count)
		if count != 1 {
			return s.makeErrJSON(400, 40001, "Challenge 不存在")
		}

		// 检查 TeamID
		s.Mysql.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: item.TeamID}}).Count(&count)
		if count != 1 {
			return s.makeErrJSON(400, 40001, "Team 不存在")
		}

		// 检查是否重复
		s.Mysql.Model(GameBox{}).Where(&GameBox{ChallengeID: item.ChallengeID, TeamID: item.TeamID}).Count(&count)
		if count != 0 {
			return s.makeErrJSON(400, 40001, "存在重复添加数据")
		}
	}

	tx := s.Mysql.Begin()
	for _, item := range inputForm {
		newGameBox := &GameBox{
			ChallengeID: item.ChallengeID,
			TeamID:      item.TeamID,
			Description: item.Description,
		}
		if tx.Create(newGameBox).RowsAffected != 1 {
			tx.Rollback()
			return s.makeErrJSON(500, 50000, "添加 GameBox 失败！")
		}
	}
	tx.Commit()
	return s.makeSuccessJSON("添加 GameBox 成功！")
}
