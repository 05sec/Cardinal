package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
)

// 攻击记录
type AttackAction struct {
	gorm.Model

	TeamID         uint // 被攻击者
	GameBoxID      uint // 被攻击者靶机
	ChallengeID    uint // 被攻击题目
	AttackerTeamID uint // 攻击者
	Round          int
}

type Flag struct {
	gorm.Model

	TeamID      uint
	GameBoxID   uint
	ChallengeID uint
	Round       int
	Flag        string
}

func (s *Service) SubmitFlag(c *gin.Context) (int, interface{}) {
	secretKey := c.GetHeader("Authorization")
	if secretKey == "" {
		return s.makeErrJSON(403, 40300, "Token 无效")
	}
	var team Team
	s.Mysql.Model(&Team{}).Where(&Team{SecretKey: secretKey}).Find(&team)
	teamID := team.ID
	if teamID == 0 {
		return s.makeErrJSON(403, 40300, "Token 无效")
	}

	type InputForm struct {
		Flag string `json:"flag" binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	var flagData Flag
	s.Mysql.Model(&Flag{}).Where(&Flag{Flag: inputForm.Flag, Round: s.Timer.NowRound}).Find(&flagData) // 注意判断是否为本轮 Flag
	if flagData.ID == 0 || teamID == flagData.TeamID {                                                 // 注意不允许提交自己的 flag
		return s.makeErrJSON(403, 40300, "Flag 错误！")
	}

	// 判断是否重复提交
	var repeatAttackCheck AttackAction
	s.Mysql.Model(&AttackAction{}).Where(&AttackAction{
		TeamID:         flagData.TeamID,
		GameBoxID:      flagData.GameBoxID,
		AttackerTeamID: teamID,
		Round:          flagData.Round,
	}).Find(&repeatAttackCheck)
	if repeatAttackCheck.ID != 0 {
		return s.makeErrJSON(403, 40301, "请勿重复提交 Flag")
	}

	// 更新靶机状态信息
	s.Mysql.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: flagData.GameBoxID}}).Update(&GameBox{IsAttacked: true})

	// 无误，加分！
	tx := s.Mysql.Begin()
	if tx.Create(&AttackAction{
		TeamID:         flagData.TeamID,
		GameBoxID:      flagData.GameBoxID,
		AttackerTeamID: teamID,
		ChallengeID:    flagData.ChallengeID,
		Round:          flagData.Round,
	}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "提交失败！")
	}
	tx.Commit()

	// 刷新总排行榜靶机状态
	s.SetRankList()

	return s.makeSuccessJSON("提交成功！")
}

func (s *Service) GetFlags(c *gin.Context) (int, interface{}) {
	pageStr := c.DefaultQuery("page", "1")
	perStr := c.DefaultQuery("per", "15")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return s.makeErrJSON(400, 40000, "page 参数错误")
	}

	per, err := strconv.Atoi(perStr)
	if err != nil || per <= 0 || per >= 100 { // 限制每页最多 100 条
		return s.makeErrJSON(400, 40001, "per 参数错误")
	}

	var total int
	s.Mysql.Model(&Flag{}).Count(&total)

	var flags []Flag
	s.Mysql.Model(&Flag{}).Offset((page - 1) * per).Limit(per).Find(&flags)

	return s.makeSuccessJSON(gin.H{
		"array": flags,
		"total": total,
	})
}

func (s *Service) GenerateFlag() (int, interface{}) {
	var gameBoxes []GameBox
	s.Mysql.Model(&GameBox{}).Find(&gameBoxes)

	// 删库
	s.Mysql.Delete(&Flag{})

	for round := 1; round <= s.Timer.TotalRound; round++ {
		// Flag = FlagPrefix + hmacSha1(TeamID + | + GameBoxID + | + Round, sha1(salt)) + FlagSuffix
		for _, gameBox := range gameBoxes {
			flag := s.Conf.FlagPrefix + s.hmacSha1Encode(fmt.Sprintf("%d|%d|%d", gameBox.TeamID, gameBox.ID, round), s.sha1Encode(s.Conf.Salt)) + s.Conf.FlagSuffix
			s.Mysql.Create(&Flag{
				TeamID:      gameBox.TeamID,
				GameBoxID:   gameBox.ID,
				ChallengeID: gameBox.ChallengeID,
				Round:       round,
				Flag:        flag,
			})
		}
	}

	return s.makeSuccessJSON("生成 Flag 成功！")
}
