package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

// AttackAction is a gorm model for database table `attack_actions`.
// Used to store the flag submitted record.
type AttackAction struct {
	gorm.Model

	TeamID         uint // Victim's team ID
	GameBoxID      uint // Victim's gamebox ID
	ChallengeID    uint // Victim's challenge ID
	AttackerTeamID uint // Attacker's Team ID
	Round          int
}

// Flag is a gorm model for database table `flags`.
// All the flags will be generated before the competition start and save in this table.
type Flag struct {
	gorm.Model

	TeamID      uint
	GameBoxID   uint
	ChallengeID uint
	Round       int
	Flag        string
}

// SubmitFlag is submit flag handler for teams.
func (s *Service) SubmitFlag(c *gin.Context) (int, interface{}) {
	// Submit flag is forbidden if the competition isn't start.
	if s.Timer.Status != "on" {
		return s.makeErrJSON(403, 40300,
			s.I18n.T(c.GetString("lang"), "general.not_begin"),
		)
	}

	secretKey := c.GetHeader("Authorization")
	if secretKey == "" {
		return s.makeErrJSON(403, 40300,
			s.I18n.T(c.GetString("lang"), "general.invalid_token"),
		)
	}
	var team Team
	s.Mysql.Model(&Team{}).Where(&Team{SecretKey: secretKey}).Find(&team)
	teamID := team.ID
	if teamID == 0 {
		return s.makeErrJSON(403, 40300,
			s.I18n.T(c.GetString("lang"), "general.invalid_token"),
		)
	}

	type InputForm struct {
		Flag string `json:"flag" binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000,
			s.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	var flagData Flag
	s.Mysql.Model(&Flag{}).Where(&Flag{Flag: inputForm.Flag, Round: s.Timer.NowRound}).Find(&flagData) // 注意判断是否为本轮 Flag
	if flagData.ID == 0 || teamID == flagData.TeamID {                                                 // 注意不允许提交自己的 flag
		return s.makeErrJSON(403, 40300,
			s.I18n.T(c.GetString("lang"), "flag.error"),
		)
	}

	// Check if the flag has been submitted by the team before.
	var repeatAttackCheck AttackAction
	s.Mysql.Model(&AttackAction{}).Where(&AttackAction{
		TeamID:         flagData.TeamID,
		GameBoxID:      flagData.GameBoxID,
		AttackerTeamID: teamID,
		Round:          flagData.Round,
	}).Find(&repeatAttackCheck)
	if repeatAttackCheck.ID != 0 {
		return s.makeErrJSON(403, 40301,
			s.I18n.T(c.GetString("lang"), "flag.repeat"),
		)
	}

	// Update the victim's gamebox status to `down`.
	s.Mysql.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: flagData.GameBoxID}}).Update(&GameBox{IsAttacked: true})

	// Save this attack record.
	tx := s.Mysql.Begin()
	if tx.Create(&AttackAction{
		TeamID:         flagData.TeamID,
		GameBoxID:      flagData.GameBoxID,
		AttackerTeamID: teamID,
		ChallengeID:    flagData.ChallengeID,
		Round:          flagData.Round,
	}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000,
			s.I18n.T(c.GetString("lang"), "flag.submit_error"),
		)
	}
	tx.Commit()

	// Update the gamebox status in ranking list.
	s.SetRankList()

	return s.makeSuccessJSON(s.I18n.T(c.GetString("lang"), "flag.submit_success"))
}

// GetFlags get flags from the database for backstage manager.
func (s *Service) GetFlags(c *gin.Context) (int, interface{}) {
	pageStr := c.DefaultQuery("page", "1")
	perStr := c.DefaultQuery("per", "15")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return s.makeErrJSON(400, 40000,
			s.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}

	per, err := strconv.Atoi(perStr)
	if err != nil || per <= 0 || per >= 100 { // 限制每页最多 100 条
		return s.makeErrJSON(400, 40001,
			s.I18n.T(c.GetString("lang"), "general.error_query"),
		)
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

// GenerateFlag is the generate flag handler for manager.
func (s *Service) GenerateFlag(c *gin.Context) (int, interface{}) {
	var gameBoxes []GameBox
	s.Mysql.Model(&GameBox{}).Find(&gameBoxes)

	startTime := time.Now().UnixNano()
	// Delete all the flags in the table.
	s.Mysql.Unscoped().Delete(&Flag{})

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

	var count int
	s.Mysql.Model(&Flag{}).Count(&count)
	endTime := time.Now().UnixNano()
	s.NewLog(WARNING, "system",
		string(s.I18n.T(c.GetString("lang"), "log.generate_flag", gin.H{"total": count, "time": float64(endTime-startTime) / float64(time.Second)})),
	)
	return s.makeSuccessJSON(s.I18n.T(c.GetString("lang"), "flag.generate_success"))
}
