package game

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vidar-team/Cardinal/internal/asteroid"
	"github.com/vidar-team/Cardinal/internal/auth/team"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/livelog"
	"github.com/vidar-team/Cardinal/internal/misc/webhook"
	"github.com/vidar-team/Cardinal/internal/timer"
	"github.com/vidar-team/Cardinal/internal/utils"
	"github.com/vidar-team/Cardinal/locales"
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
func CheckDown(c *gin.Context) (int, interface{}) {
	// Check down is forbidden if the competition isn't start.
	if timer.Get().Status != "on" {
		return utils.MakeErrJSON(403, 40310,
			locales.I18n.T(c.GetString("lang"), "general.not_begin"),
		)
	}

	type InputForm struct {
		GameBoxID uint `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40026,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	// Does it check down one gamebox repeatedly in one round?
	var repeatCheck DownAction
	db.MySQL.Model(&DownAction{}).Where(&DownAction{
		GameBoxID: inputForm.GameBoxID,
		Round:     timer.Get().NowRound,
	}).Find(&repeatCheck)
	if repeatCheck.ID != 0 {
		return utils.MakeErrJSON(403, 40311,
			locales.I18n.T(c.GetString("lang"), "check.repeat"),
		)
	}

	// Check the gamebox is existed or not.
	var gameBox GameBox
	db.MySQL.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: inputForm.GameBoxID}}).Find(&gameBox)
	if gameBox.ID == 0 {
		return utils.MakeErrJSON(403, 40312,
			locales.I18n.T(c.GetString("lang"), "gamebox.not_found"),
		)
	}
	if !gameBox.Visible {
		return utils.MakeErrJSON(403, 40314,
			locales.I18n.T(c.GetString("lang"), "check.not_visible"),
		)
	}

	// No problem! Update the gamebox status to down.
	db.MySQL.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: gameBox.ID}}).Update(&GameBox{IsDown: true})

	tx := db.MySQL.Begin()
	if tx.Create(&DownAction{
		TeamID:      gameBox.TeamID,
		ChallengeID: gameBox.ChallengeID,
		GameBoxID:   inputForm.GameBoxID,
		Round:       timer.Get().NowRound,
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50015,
			locales.I18n.T(c.GetString("lang"), "general.server_error"),
		)
	}
	tx.Commit()

	// Check down hook
	go webhook.Add(webhook.CHECK_DOWN_HOOK, gin.H{"team": gameBox.TeamID, "gamebox": gameBox.ID})

	// Update the gamebox status in ranking list.
	SetRankList()

	// Asteroid Unity3D
	asteroid.SendStatus(int(gameBox.TeamID), "down")

	var t team.Team
	db.MySQL.Model(&team.Team{}).Where(&team.Team{Model: gorm.Model{ID: gameBox.TeamID}}).Find(&t)
	var challenge Challenge
	db.MySQL.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: gameBox.ChallengeID}}).Find(&challenge)
	// Live log
	_ = livelog.Stream.Write(livelog.GlobalStream, livelog.NewLine("check_down",
		gin.H{"Team": t.Name, "Challenge": challenge.Title}))

	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}
