package game

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/utils"
)

// SetVisible is setting challenge visible status handler.
// When a challenge's visible status changed, all the teams' challenge scores and their total scores will be calculated immediately.
// The ranking list will also be updated.
func SetVisible(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID      uint `binding:"required"`
		Visible bool
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40027,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	var checkChallenge db.Challenge
	db.MySQL.Where(&db.Challenge{Model: gorm.Model{ID: inputForm.ID}}).Find(&checkChallenge)
	if checkChallenge.Title == "" {
		return utils.MakeErrJSON(404, 40402,
			locales.I18n.T(c.GetString("lang"), "challenge.not_found"),
		)
	}

	db.MySQL.Model(&db.GameBox{}).Where("challenge_id = ?", inputForm.ID).Update(map[string]interface{}{"visible": inputForm.Visible})

	// Calculate all the teams' score. (Only visible challenges)
	calculateTeamScore()
	// Refresh the ranking list table's header.
	SetRankListTitle()
	// Refresh the ranking list teams' scores.
	SetRankList()

	status := "invisible"
	if inputForm.Visible {
		status = "visible"
	}
	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.set_challenge_"+status, gin.H{"challenge": checkChallenge.Title})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "gamebox.visibility_success"))
}

// GetAllChallenges get all challenges from the database.
func GetAllChallenges(c *gin.Context) (int, interface{}) {
	var challenges []db.Challenge
	db.MySQL.Model(&db.Challenge{}).Find(&challenges)
	type resultStruct struct {
		ID              uint
		CreatedAt       time.Time
		Title           string
		Visible         bool
		BaseScore       int
		AutoRefreshFlag bool
		Command         string
	}

	var res []resultStruct
	for _, v := range challenges {
		// For the challenge model doesn't have the `visible` field,
		// We can only get the challenge's visible status by one of its gamebox.
		// TODO: Need to find a better way to get the challenge's visible status.
		var gameBox db.GameBox
		db.MySQL.Where(&db.GameBox{ChallengeID: v.ID}).Limit(1).Find(&gameBox)

		res = append(res, resultStruct{
			ID:              v.ID,
			CreatedAt:       v.CreatedAt,
			Title:           v.Title,
			Visible:         gameBox.Visible,
			BaseScore:       v.BaseScore,
			AutoRefreshFlag: v.AutoRefreshFlag,
			Command:         v.Command,
		})
	}
	return utils.MakeSuccessJSON(res)
}

// NewChallenge is new challenge handler for manager.
func NewChallenge(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Title           string `binding:"required"`
		BaseScore       int    `binding:"required"`
		AutoRefreshFlag bool
		Command         string
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40028,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	if inputForm.AutoRefreshFlag && inputForm.Command == "" {
		return utils.MakeErrJSON(400, 40029,
			locales.I18n.T(c.GetString("lang"), "challenge.empty_command"))
	}

	if !inputForm.AutoRefreshFlag {
		inputForm.Command = ""
	}

	newChallenge := &db.Challenge{
		Title:           inputForm.Title,
		BaseScore:       inputForm.BaseScore,
		AutoRefreshFlag: inputForm.AutoRefreshFlag,
		Command:         inputForm.Command,
	}
	var checkChallenge db.Challenge

	db.MySQL.Model(&db.Challenge{}).Where(&db.Challenge{Title: newChallenge.Title}).Find(&checkChallenge)
	if checkChallenge.Title != "" {
		return utils.MakeErrJSON(403, 40313,
			locales.I18n.T(c.GetString("lang"), "general.post_repeat"),
		)
	}

	tx := db.MySQL.Begin()
	if tx.Create(newChallenge).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50016,
			locales.I18n.T(c.GetString("lang"), "challenge.post_error"),
		)
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.new_challenge", gin.H{"title": newChallenge.Title})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "challenge.post_success"))
}

// EditChallenge is edit challenge handler for manager.
func EditChallenge(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID              uint   `binding:"required"`
		Title           string `binding:"required"`
		BaseScore       int    `binding:"required"`
		AutoRefreshFlag bool
		Command         string
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40028,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	if inputForm.AutoRefreshFlag && inputForm.Command == "" {
		return utils.MakeErrJSON(400, 40029,
			locales.I18n.T(c.GetString("lang"), "challenge.empty_command"))
	}

	// True off auto refresh flag, clean the command.
	if !inputForm.AutoRefreshFlag {
		inputForm.Command = ""
	}

	var checkChallenge db.Challenge
	db.MySQL.Where(&db.Challenge{Model: gorm.Model{ID: inputForm.ID}}).Find(&checkChallenge)
	if checkChallenge.Title == "" {
		return utils.MakeErrJSON(404, 40403,
			locales.I18n.T(c.GetString("lang"), "challenge.not_found"),
		)
	}

	// For the `AutoRefreshFlag` is a boolean value, use map here.
	editChallenge := map[string]interface{}{
		"Title":           inputForm.Title,
		"BaseScore":       inputForm.BaseScore,
		"AutoRefreshFlag": inputForm.AutoRefreshFlag,
		"Command":         inputForm.Command,
	}
	tx := db.MySQL.Begin()
	if tx.Model(&db.Challenge{}).Where(&db.Challenge{Model: gorm.Model{ID: inputForm.ID}}).Updates(editChallenge).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50017,
			locales.I18n.T(c.GetString("lang"), "challenge.put_error"),
		)
	}
	tx.Commit()

	// If the challenge's score is updated, we need to calculate the gameboxes' scores and the teams' scores.
	if inputForm.BaseScore != checkChallenge.BaseScore {
		// Calculate all the teams' score. (Only visible challenges)
		calculateTeamScore()
		// Refresh the ranking list table's header.
		SetRankListTitle()
		// Refresh the ranking list teams' scores.
		SetRankList()
	}

	// If the challenge's title is updated, we just need to update the ranking list table's header.
	if inputForm.Title != checkChallenge.Title {
		SetRankListTitle()
	}

	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "challenge.put_success"))
}

// DeleteChallenge is delete challenge handler for manager.
func DeleteChallenge(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40030,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40030,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	var challenge db.Challenge
	db.MySQL.Where(&db.Challenge{Model: gorm.Model{ID: uint(id)}}).Find(&challenge)
	if challenge.Title == "" {
		return utils.MakeErrJSON(404, 40403,
			locales.I18n.T(c.GetString("lang"), "challenge.not_found"),
		)
	}

	tx := db.MySQL.Begin()
	// 同时删除 GameBox
	tx.Where("challenge_id = ?", uint(id)).Delete(&db.GameBox{})
	if tx.Where("id = ?", uint(id)).Delete(&db.Challenge{}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50018,
			locales.I18n.T(c.GetString("lang"), "challenge.delete_error"),
		)
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.delete_challenge", gin.H{"title": challenge.Title})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "challenge.delete_success"))
}
