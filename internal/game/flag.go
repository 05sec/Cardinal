package game

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/asteroid"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/dynamic_config"
	"github.com/vidar-team/Cardinal/internal/livelog"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/misc/webhook"
	"github.com/vidar-team/Cardinal/internal/timer"
	"github.com/vidar-team/Cardinal/internal/utils"
)

// SubmitFlag is submit flag handler for teams.
func SubmitFlag(c *gin.Context) (int, interface{}) {
	// Submit flag is forbidden if the competition isn't start.
	if timer.Get().Status != "on" {
		return utils.MakeErrJSON(403, 40304,
			locales.I18n.T(c.GetString("lang"), "general.not_begin"),
		)
	}

	secretKey := c.GetHeader("Authorization")
	if secretKey == "" {
		return utils.MakeErrJSON(403, 40305,
			locales.I18n.T(c.GetString("lang"), "general.invalid_token"),
		)
	}
	var t db.Team
	db.MySQL.Model(&db.Team{}).Where(&db.Team{SecretKey: secretKey}).Find(&t)
	teamID := t.ID
	if teamID == 0 {
		return utils.MakeErrJSON(403, 40306,
			locales.I18n.T(c.GetString("lang"), "general.invalid_token"),
		)
	}

	type InputForm struct {
		Flag string `json:"flag" binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40021,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	// Remove the space
	inputForm.Flag = strings.TrimSpace(inputForm.Flag)

	var flagData db.Flag
	db.MySQL.Model(&db.Flag{}).Where(&db.Flag{Flag: inputForm.Flag, Round: timer.Get().NowRound}).Find(&flagData) // 注意判断是否为本轮 Flag
	if flagData.ID == 0 || teamID == flagData.TeamID {                                                            // 注意不允许提交自己的 flag
		return utils.MakeErrJSON(403, 40307,
			locales.I18n.T(c.GetString("lang"), "flag.wrong"),
		)
	}

	// Check the challenge is visible or not.
	var gamebox db.GameBox
	db.MySQL.Model(&db.GameBox{}).Where(&db.GameBox{Model: gorm.Model{ID: flagData.GameBoxID}, Visible: true}).Find(&gamebox)
	if gamebox.ID == 0 {
		return utils.MakeErrJSON(403, 40308,
			locales.I18n.T(c.GetString("lang"), "flag.wrong"),
		)
	}

	// Check if the flag has been submitted by the team before.
	var repeatAttackCheck db.AttackAction
	db.MySQL.Model(&db.AttackAction{}).Where(&db.AttackAction{
		TeamID:         flagData.TeamID,
		GameBoxID:      flagData.GameBoxID,
		AttackerTeamID: teamID,
		Round:          flagData.Round,
	}).Find(&repeatAttackCheck)
	if repeatAttackCheck.ID != 0 {
		// Animate Asteroid
		animateAsteroid, _ := strconv.ParseBool(dynamic_config.Get(utils.ANIMATE_ASTEROID))
		if animateAsteroid {
			asteroid.SendAttack(int(teamID), int(flagData.TeamID))
		}

		return utils.MakeErrJSON(403, 40309,
			locales.I18n.T(c.GetString("lang"), "flag.repeat"),
		)
	}

	// Update the victim's gamebox status to `down`.
	db.MySQL.Model(&db.GameBox{}).Where(&db.GameBox{Model: gorm.Model{ID: flagData.GameBoxID}}).Update(&db.GameBox{IsAttacked: true})

	// Save this attack record.
	tx := db.MySQL.Begin()
	if tx.Create(&db.AttackAction{
		TeamID:         flagData.TeamID,
		GameBoxID:      flagData.GameBoxID,
		AttackerTeamID: teamID,
		ChallengeID:    flagData.ChallengeID,
		Round:          flagData.Round,
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50013,
			locales.I18n.T(c.GetString("lang"), "flag.submit_error"),
		)
	}
	tx.Commit()

	// Update the gamebox status in ranking list.
	SetRankList()
	// Webhook
	go webhook.Add(webhook.SUBMIT_FLAG_HOOK, gin.H{"from": teamID, "to": gamebox.TeamID, "gamebox": gamebox.ID})
	// Send Unity3D attack message.
	asteroid.SendAttack(int(teamID), int(flagData.TeamID))

	// Get attack team data
	var flagTeam db.Team
	db.MySQL.Model(&db.Team{}).Where(&db.Team{Model: gorm.Model{ID: flagData.TeamID}}).Find(&flagTeam)
	// Get challenge data
	var challenge db.Challenge
	db.MySQL.Model(&db.Challenge{}).Where(&db.Challenge{Model: gorm.Model{ID: flagData.ChallengeID}}).Find(&challenge)
	// Live log
	_ = livelog.Stream.Write(livelog.GlobalStream, livelog.NewLine("submit_flag",
		gin.H{"From": t.Name, "To": flagTeam.Name, "Challenge": challenge.Title}))

	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "flag.submit_success"))
}

// GetFlags get flags from the database for backstage manager.
func GetFlags(c *gin.Context) (int, interface{}) {
	pageStr := c.DefaultQuery("page", "1")
	perStr := c.DefaultQuery("per", "15")

	// filter
	roundStr := c.DefaultQuery("round", "0")
	teamStr := c.DefaultQuery("team", "0")
	challengeStr := c.DefaultQuery("challenge", "0")

	round, err := strconv.Atoi(roundStr)
	if err != nil || round < 0 {
		return utils.MakeErrJSON(400, 40022,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	teamID, err := strconv.Atoi(teamStr)
	if err != nil || teamID < 0 {
		return utils.MakeErrJSON(400, 40022,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	challengeID, err := strconv.Atoi(challengeStr)
	if err != nil || challengeID < 0 {
		return utils.MakeErrJSON(400, 40022,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return utils.MakeErrJSON(400, 40022,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}

	per, err := strconv.Atoi(perStr)
	if err != nil || per <= 0 || per >= 100 { // 限制每页最多 100 条
		return utils.MakeErrJSON(400, 40023,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}

	var total int
	db.MySQL.Model(&db.Flag{}).Where(&db.Flag{
		TeamID:      uint(teamID),
		ChallengeID: uint(challengeID),
		Round:       round,
	}).Count(&total)

	var flags []db.Flag
	db.MySQL.Model(&db.Flag{}).Where(&db.Flag{
		TeamID:      uint(teamID),
		ChallengeID: uint(challengeID),
		Round:       round,
	}).Offset((page - 1) * per).Limit(per).Find(&flags)

	return utils.MakeSuccessJSON(gin.H{
		"array": flags,
		"total": total,
	})
}

// ExportFlag exports the flags of a challenge.
func ExportFlag(c *gin.Context) (int, interface{}) {
	challengeIDStr := c.DefaultQuery("id", "1")

	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil || challengeID <= 0 {
		return utils.MakeErrJSON(400, 40024,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}

	var flags []db.Flag
	db.MySQL.Model(&db.Flag{}).Where(&db.Flag{ChallengeID: uint(challengeID)}).Find(&flags)
	return utils.MakeSuccessJSON(flags)
}

// GenerateFlag is the generate flag handler for manager.
func GenerateFlag(c *gin.Context) (int, interface{}) {
	var gameBoxes []db.GameBox
	db.MySQL.Model(&db.GameBox{}).Find(&gameBoxes)

	startTime := time.Now().UnixNano()
	// Delete all the flags in the table.
	db.MySQL.Unscoped().Delete(&db.Flag{})

	flagPrefix := dynamic_config.Get(utils.FLAG_PREFIX_CONF)
	flagSuffix := dynamic_config.Get(utils.FLAG_SUFFIX_CONF)

	salt := utils.Sha1Encode(conf.Get().Salt)
	for round := 1; round <= timer.Get().TotalRound; round++ {
		// Flag = FlagPrefix + hmacSha1(TeamID + | + GameBoxID + | + Round, sha1(salt)) + FlagSuffix
		for _, gameBox := range gameBoxes {
			flag := flagPrefix + utils.HmacSha1Encode(fmt.Sprintf("%d|%d|%d", gameBox.TeamID, gameBox.ID, round), salt) + flagSuffix
			db.MySQL.Create(&db.Flag{
				TeamID:      gameBox.TeamID,
				GameBoxID:   gameBox.ID,
				ChallengeID: gameBox.ChallengeID,
				Round:       round,
				Flag:        flag,
			})
		}
	}

	var count int
	db.MySQL.Model(&db.Flag{}).Count(&count)
	endTime := time.Now().UnixNano()
	logger.New(logger.WARNING, "system",
		string(locales.I18n.T(c.GetString("lang"), "log.generate_flag", gin.H{"total": count, "time": float64(endTime-startTime) / float64(time.Second)})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "flag.generate_success"))
}

// RefreshFlag refreshes all the flags in current round.
func RefreshFlag() {
	// Get the auto refresh flag challenges.
	var challenges []db.Challenge
	db.MySQL.Model(&db.Challenge{}).Where(&db.Challenge{AutoRefreshFlag: true}).Find(&challenges)

	for _, challenge := range challenges {
		var gameboxes []db.GameBox
		db.MySQL.Model(&db.GameBox{}).Where(&db.GameBox{ChallengeID: challenge.ID}).Find(&gameboxes)

		for _, gamebox := range gameboxes {
			go func(gamebox db.GameBox, challenge db.Challenge) {
				var flag db.Flag
				db.MySQL.Model(&db.Flag{}).Where(&db.Flag{
					TeamID:    gamebox.TeamID,
					GameBoxID: gamebox.ID,
					Round:     timer.Get().NowRound,
				}).Find(&flag)
				// Replace the flag placeholder.
				// strings.ReplaceAll need Go 1.13+, so we use strings.Replace here.
				command := strings.Replace(challenge.Command, "{{FLAG}}", flag.Flag, -1)
				_, err := utils.SSHExecute(gamebox.IP, gamebox.SSHPort, gamebox.SSHUser, gamebox.SSHPassword, command)
				if err != nil {
					logger.New(logger.IMPORTANT, "ssh_error", fmt.Sprintf("Team:%d Gamebox:%d Round:%d SSH 更新 Flag 失败：%v", gamebox.TeamID, gamebox.ID, timer.Get().NowRound, err.Error()))
				}

			}(gamebox, challenge)
		}
	}
}

func TestAllSSH(c *gin.Context) (int, interface{}) {
	var challenges []db.Challenge
	db.MySQL.Model(&db.Challenge{}).Where(&db.Challenge{AutoRefreshFlag: true}).Find(&challenges)

	type errorMessage struct {
		TeamID      uint
		ChallengeID uint
		GameBoxID   uint
		Error       string
	}
	var errs []errorMessage

	wg := sync.WaitGroup{}
	for _, challenge := range challenges {
		var gameboxes []db.GameBox
		db.MySQL.Model(&db.GameBox{}).Where(&db.GameBox{ChallengeID: challenge.ID}).Find(&gameboxes)

		for _, gamebox := range gameboxes {
			wg.Add(1)
			go func(gamebox db.GameBox, challenge db.Challenge) {
				defer wg.Done()
				_, err := utils.SSHExecute(gamebox.IP, gamebox.SSHPort, gamebox.SSHUser, gamebox.SSHPassword, "whoami")
				if err != nil {
					errs = append(errs, errorMessage{
						TeamID:      gamebox.TeamID,
						ChallengeID: challenge.ID,
						GameBoxID:   gamebox.ID,
						Error:       err.Error(),
					})
				}
			}(gamebox, challenge)
		}
	}
	wg.Wait()
	return utils.MakeSuccessJSON(errs)
}

func TestSSH(c *gin.Context) (int, interface{}) {
	var inputForm struct {
		IP       string `binding:"required"`
		Port     string `binding:"required"`
		User     string `binding:"required"`
		Password string `binding:"required"`
		Command  string `binding:"required"`
	}
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40036,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	output, err := utils.SSHExecute(inputForm.IP, inputForm.Port, inputForm.User, inputForm.Password, inputForm.Command)
	if err != nil {
		return utils.MakeErrJSON(400, 40037, err)
	}
	return utils.MakeSuccessJSON(output)
}

func GetLatestScoreRound() int {
	var latestScore db.Score
	db.MySQL.Model(&db.Score{}).Order("`round` DESC").Limit(1).Find(&latestScore)
	return latestScore.Round
}
