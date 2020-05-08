package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vidar-team/Cardinal/src/locales"
	"github.com/vidar-team/Cardinal/src/utils"
	"math"
	"strconv"
)

// GameBox is a gorm model for database table `gameboxes`.
type GameBox struct {
	gorm.Model
	ChallengeID uint
	TeamID      uint

	IP          string
	Port        string
	SSHPort     string
	SSHUser     string
	SSHPassword string
	Description string
	Visible     bool
	Score       float64 // The score can be negative.
	IsDown      bool
	IsAttacked  bool
}

// GetSelfGameBoxes returns the gameboxes which belong to the team.
func (s *Service) GetSelfGameBoxes(c *gin.Context) (int, interface{}) {
	if s.Timer.Status == "wait" {
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

	s.Mysql.Table("game_boxes").Where(&GameBox{TeamID: teamID.(uint), Visible: true}).Order("challenge_id").Find(&gameBoxes)
	for index, gameBox := range gameBoxes {
		var challenge Challenge
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: gameBox.ChallengeID}}).Find(&challenge)
		gameBoxes[index].Title = challenge.Title
	}
	return utils.MakeSuccessJSON(gameBoxes)
}

// GetGameBoxes returns the gameboxes for manager.
func (s *Service) GetGameBoxes(c *gin.Context) (int, interface{}) {
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
	s.Mysql.Model(&GameBox{}).Count(&total)
	var gameBox []GameBox
	s.Mysql.Model(&GameBox{}).Offset((page - 1) * perPage).Limit(perPage).Find(&gameBox)

	return utils.MakeSuccessJSON(gin.H{
		"Data":      gameBox,
		"Total":     total,
		"TotalPage": math.Ceil(float64(total / perPage)),
	})
}

// NewGameBoxes is add a new gamebox handler for manager.
func (s *Service) NewGameBoxes(c *gin.Context) (int, interface{}) {
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
		var challenge Challenge
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: item.ChallengeID}}).Find(&challenge)
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
		s.Mysql.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: item.TeamID}}).Count(&count)
		if count != 1 {
			return utils.MakeErrJSON(400, 40018,
				locales.I18n.T(c.GetString("lang"), "team.not_found"),
			)
		}

		// Check if the gamebox is existed by challenge ID and team ID,
		// since every team should have only one gamebox for each challenge.
		s.Mysql.Model(GameBox{}).Where(&GameBox{ChallengeID: item.ChallengeID, TeamID: item.TeamID}).Count(&count)
		if count != 0 {
			return utils.MakeErrJSON(400, 40019,
				locales.I18n.T(c.GetString("lang"), "gamebox.repeat"),
			)
		}
	}

	tx := s.Mysql.Begin()
	for _, item := range inputForm {
		newGameBox := &GameBox{
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

	s.NewLog(NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.new_gamebox", gin.H{"count": len(inputForm)})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "gamebox.post_success"))
}

// EditGameBox is edit gamebox handler for manager.
func (s *Service) EditGameBox(c *gin.Context) (int, interface{}) {
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

	tx := s.Mysql.Begin()
	if tx.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: inputForm.ID}}).Updates(&GameBox{
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
