package game

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/vidar-team/Cardinal/internal/dbold"
	"github.com/vidar-team/Cardinal/internal/dynamic_config"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/timer"
	"github.com/vidar-team/Cardinal/internal/utils"
)

// GetSelfGameBoxes returns the gameboxes which belong to the team.
func GetSelfGameBoxes(c *gin.Context) (int, interface{}) {
	if timer.Get().Status == "wait" {
		return utils.MakeSuccessJSON([]int{})
	}

	var gameBoxes []struct {
		ChallengeID uint
		Title       string
		IP          string
		Port        string
		Description string
		Score       float64
		IsDown      bool
		IsAttacked  bool
	}
	teamID, _ := c.Get("teamID")

	dbold.MySQL.Table("game_boxes").Where(&dbold.GameBox{TeamID: teamID.(uint), Visible: true}).Order("challenge_id").Find(&gameBoxes)
	for index, gameBox := range gameBoxes {
		var challenge dbold.Challenge
		dbold.MySQL.Model(&dbold.Challenge{}).Where(&dbold.Challenge{Model: gorm.Model{ID: gameBox.ChallengeID}}).Find(&challenge)
		gameBoxes[index].Title = challenge.Title
	}
	return utils.MakeSuccessJSON(gameBoxes)
}

// GetGameBoxes returns the gameboxes for manager.
func GetGameBoxes(c *gin.Context) (int, interface{}) {
	pageStr := c.Query("page")
	perPageStr := c.Query("per")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return utils.MakeErrJSON(400, 40013,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage <= 0 {
		return utils.MakeErrJSON(400, 40014,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}

	var total int
	dbold.MySQL.Model(&dbold.GameBox{}).Count(&total)
	var gameBox []dbold.GameBox
	dbold.MySQL.Model(&dbold.GameBox{}).Offset((page - 1) * perPage).Limit(perPage).Find(&gameBox)

	return utils.MakeSuccessJSON(gin.H{
		"Data":      gameBox,
		"Total":     total,
		"TotalPage": int(math.Ceil(float64(total) / float64(perPage))),
	})
}

// NewGameBoxes is add a new gamebox handler for manager.
func NewGameBoxes(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ChallengeID uint   `binding:"required"`
		TeamID      uint   `binding:"required"`
		IP          string `binding:"required"`
		Port        string `binding:"required"`
		SSHPort     string
		SSHUser     string
		SSHPassword string
		Description string `binding:"required"`

		Score float64 // not for form
	}
	var inputForm []*InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40015,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}

	for _, item := range inputForm {
		var count int

		// Check the ChallengeID
		var challenge dbold.Challenge
		dbold.MySQL.Model(&dbold.Challenge{}).Where(&dbold.Challenge{Model: gorm.Model{ID: item.ChallengeID}}).Find(&challenge)
		if challenge.ID == 0 {
			return utils.MakeErrJSON(400, 40016,
				locales.I18n.T(c.GetString("lang"), "challenge.not_found"),
			)
		}
		// Set the default score.
		item.Score = float64(challenge.BaseScore)

		// Check SSH config
		if challenge.AutoRefreshFlag {
			if item.SSHPort == "" || item.SSHUser == "" || item.SSHPassword == "" {
				return utils.MakeErrJSON(400, 40017,
					locales.I18n.T(c.GetString("lang"), "gamebox.auto_refresh_flag_error"),
				)
			}
		}

		// Check the TeamID
		dbold.MySQL.Model(&dbold.Team{}).Where(&dbold.Team{Model: gorm.Model{ID: item.TeamID}}).Count(&count)
		if count != 1 {
			return utils.MakeErrJSON(400, 40018,
				locales.I18n.T(c.GetString("lang"), "team.not_found"),
			)
		}

		// Check if the gamebox is existed by challenge ID and team ID,
		// since every team should have only one gamebox for each challenge.
		dbold.MySQL.Model(dbold.GameBox{}).Where(&dbold.GameBox{ChallengeID: item.ChallengeID, TeamID: item.TeamID}).Count(&count)
		if count != 0 {
			return utils.MakeErrJSON(400, 40019,
				locales.I18n.T(c.GetString("lang"), "gamebox.repeat"),
			)
		}
	}

	tx := dbold.MySQL.Begin()
	for _, item := range inputForm {
		newGameBox := &dbold.GameBox{
			ChallengeID: item.ChallengeID,
			TeamID:      item.TeamID,
			IP:          item.IP,
			Port:        item.Port,
			SSHPort:     item.SSHPort,
			SSHUser:     item.SSHUser,
			SSHPassword: item.SSHPassword,
			Score:       item.Score,
			Description: item.Description,
		}
		if tx.Create(newGameBox).RowsAffected != 1 {
			tx.Rollback()
			return utils.MakeErrJSON(500, 50011,
				locales.I18n.T(c.GetString("lang"), "gamebox.post_error"),
			)
		}
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.new_gamebox", gin.H{"count": len(inputForm)})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "gamebox.post_success"))
}

// EditGameBox is edit gamebox handler for manager.
func EditGameBox(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID uint `binding:"required"`

		IP          string `binding:"required"`
		Port        string `binding:"required"`
		SSHPort     string
		SSHUser     string
		SSHPassword string
		Description string `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40020,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	tx := dbold.MySQL.Begin()
	if tx.Model(&dbold.GameBox{}).Where(&dbold.GameBox{Model: gorm.Model{ID: inputForm.ID}}).Updates(&dbold.GameBox{
		IP:          inputForm.IP,
		Port:        inputForm.Port,
		SSHPort:     inputForm.SSHPort,
		SSHUser:     inputForm.SSHUser,
		SSHPassword: inputForm.SSHPassword,
		Description: inputForm.Description,
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50012,
			locales.I18n.T(c.GetString("lang"), "gamebox.put_error"),
		)
	}
	tx.Commit()

	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "gamebox.put_success"))
}

// GetOthersGameBox returns the other teams' gameboxes if the config is set to true.
func GetOthersGameBox(c *gin.Context) (int, interface{}) {
	animateAsteroid, _ := strconv.ParseBool(dynamic_config.Get(utils.SHOW_OTHERS_GAMEBOX))
	if !animateAsteroid {
		return utils.MakeErrJSON(400, 40047, locales.I18n.T(c.GetString("lang"), "general.no_auth"))
	}
	var gameboxs []struct {
		ID            uint
		ChallengeName string
		TeamName      string
		IP            string
		Port          string
	}
	dbold.MySQL.Raw("SELECT `game_boxes`.`id`, `game_boxes`.`ip`,`game_boxes`.`port`, `challenges`.`title` AS challenge_name, `teams`.`name` AS team_name " +
		"FROM `game_boxes`, `teams`, `challenges` " +
		"WHERE `game_boxes`.`visible` = 1 AND `game_boxes`.`challenge_id` = `challenges`.`id` AND `game_boxes`.`team_id` = `teams`.`id`").Scan(&gameboxs)
	return utils.MakeSuccessJSON(gameboxs)
}

func CleanGameBoxStatus() {
	dbold.MySQL.Model(&dbold.GameBox{}).Update(map[string]interface{}{"is_down": false, "is_attacked": false})
}

func ResetAllGameBoxes(c *gin.Context) (int, interface{}) {
	dbold.MySQL.Model(&dbold.AttackAction{}).Delete(&dbold.AttackAction{})
	dbold.MySQL.Model(&dbold.DownAction{}).Delete(&dbold.DownAction{})

	CleanGameBoxStatus()
	SetRankList()

	logger.New(logger.IMPORTANT, "manager_operate", string(locales.I18n.T(c.GetString("lang"), "gamebox.reset_success")))
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "gamebox.reset_success"))
}
