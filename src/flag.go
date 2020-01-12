package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
)

// 攻击记录
type AttackAction struct {
	gorm.Model

	TeamID         uint // 被攻击者
	GameBoxID      uint // 被攻击者靶机
	AttackerTeamID uint // 攻击者
	Round          int
}

type Flag struct {
	gorm.Model

	TeamID    uint
	GameBoxID uint
	Round     int
	Flag      string
}

func (s *Service) SubmitFlag(c *gin.Context) (int, interface{}) {
	teamID := c.GetInt("teamID")
	if teamID == 0 {
		return s.makeErrJSON(500, 50001, "Server error")
	}
	type InputForm struct {
		Flag string `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	var flagData Flag
	s.Mysql.Model(&Flag{}).Where(&Flag{Flag: inputForm.Flag, Round: s.Timer.NowRound}).Find(&flagData) // 注意判断是否为本轮 Flag
	if flagData.ID == 0 {
		return s.makeErrJSON(403, 40300, "Flag 错误！")
	}

	// 判断是否重复提交
	var repeatAttackCheck AttackAction
	s.Mysql.Model(&AttackAction{}).Where(&AttackAction{
		TeamID:         flagData.TeamID,
		GameBoxID:      flagData.GameBoxID,
		AttackerTeamID: uint(teamID),
		Round:          flagData.Round,
	}).Find(&repeatAttackCheck)
	if repeatAttackCheck.ID != 0 {
		return s.makeErrJSON(403, 40301, "请勿重复提交 Flag")
	}

	// 无误，加分！
	tx := s.Mysql.Begin()
	if tx.Create(&AttackAction{
		TeamID:         flagData.TeamID,
		GameBoxID:      flagData.GameBoxID,
		AttackerTeamID: uint(teamID),
		Round:          flagData.Round,
	}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "提交失败！")
	}
	tx.Commit()
	return s.makeSuccessJSON("提交成功！")
}

func (s *Service) GenerateFlag() (int, interface{}) {
	var gameBoxes []GameBox
	s.Mysql.Model(&GameBox{}).Find(&gameBoxes)

	// 删库
	s.Mysql.Delete(&Flag{})

	for round := 1; round <= s.Timer.TotalRound; round++ {
		// Flag = FlagPrefix + sha1(TeamID + GameBoxID + sha1(Salt) + Round) + FlagSuffix
		for _, gameBox := range gameBoxes {
			flag := s.Conf.FlagPrefix + s.sha1Encode(strconv.Itoa(int(gameBox.TeamID))+strconv.Itoa(int(gameBox.ID))+s.sha1Encode(s.Conf.Salt)+strconv.Itoa(round)) + s.Conf.FlagSuffix
			s.Mysql.Create(&Flag{
				TeamID:    gameBox.TeamID,
				GameBoxID: gameBox.ID,
				Round:     round,
				Flag:      flag,
			})
		}
	}

	return s.makeSuccessJSON("生成 Flag 成功！")
}
