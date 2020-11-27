package team

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/thanhpk/randstr"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/utils"
)

// TeamLogin is the team login handler.
func TeamLogin(c *gin.Context) (int, interface{}) {
	type TeamLoginForm struct {
		Name     string `json:"Name" binding:"required"`
		Password string `json:"Password" binding:"required"`
	}

	var formData TeamLoginForm
	err := c.BindJSON(&formData)
	if err != nil {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	var team db.Team
	db.MySQL.Where(&db.Team{Name: formData.Name}).Find(&team)
	if team.Name != "" && utils.CheckPassword(formData.Password, team.Password) {
		// Login successfully
		token := utils.GenerateToken()

		tx := db.MySQL.Begin()
		if tx.Create(&db.Token{TeamID: team.ID, Token: token}).RowsAffected != 1 {
			tx.Rollback()
			return utils.MakeErrJSON(500, 50000,
				locales.I18n.T(c.GetString("lang"), "general.server_error"),
			)
		}
		tx.Commit()
		return utils.MakeSuccessJSON(token)
	}
	return utils.MakeErrJSON(403, 40301,
		locales.I18n.T(c.GetString("lang"), "team.login_error"),
	)
}

// TeamLogout is the team logout handler.
func TeamLogout(c *gin.Context) (int, interface{}) {
	token := c.GetHeader("Authorization")
	if token != "" {
		db.MySQL.Model(&db.Token{}).Where("token = ?", token).Delete(&db.Token{})
	}
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "team.logout_success"))
}

// GetTeamInfo returns the team its info.
func GetTeamInfo(c *gin.Context) (int, interface{}) {
	teamID, ok := c.Get("teamID")
	if !ok {
		return utils.MakeErrJSON(500, 50001,
			locales.I18n.T(c.GetString("lang"), "general.server_error"),
		)
	}

	var teamInfo db.Team
	rank := 0
	var teams []db.Team

	db.MySQL.Model(&db.Team{}).Order("`score` DESC").Find(&teams)
	// Get the team rank by its index.
	for index, t := range teams {
		if teamID.(uint) == t.ID {
			rank = index + 1
			teamInfo = t
			break
		}
	}

	return utils.MakeSuccessJSON(gin.H{
		"Name":  teamInfo.Name,
		"Logo":  teamInfo.Logo,
		"Score": teamInfo.Score,
		"Rank":  rank,
		"Token": teamInfo.SecretKey,
	})
}

// GetAllTeams returns all the teams info for manager.
func GetAllTeams(c *gin.Context) (int, interface{}) {
	var teams []db.Team
	db.MySQL.Model(&db.Team{}).Find(&teams)
	return utils.MakeSuccessJSON(teams)
}

// NewTeams is add new team(s) handler.
func NewTeams(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Name string `binding:"required"`
		Logo string `binding:"required"`
	}
	var inputForm []InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40001,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	// Check if the team name repeat in the form.
	tmpTeamName := make(map[string]int)
	for _, item := range inputForm {
		tmpTeamName[item.Name] = 0

		// Check if the team name repeat in the database.
		var count int
		db.MySQL.Model(db.Team{}).Where(&db.Team{Name: item.Name}).Count(&count)
		if count != 0 {
			return utils.MakeErrJSON(400, 40002,
				locales.I18n.T(c.GetString("lang"), "team.repeat"),
			)
		}
		// Team name can't be empty.
		if item.Name == "" {
			return utils.MakeErrJSON(400, 40003,
				locales.I18n.T(c.GetString("lang"), "team.team_name_empty"),
			)
		}
	}
	if len(tmpTeamName) != len(inputForm) {
		return utils.MakeErrJSON(400, 40004,
			locales.I18n.T(c.GetString("lang"), "team.repeat"),
		)
	}

	type resultItem struct {
		Name     string
		Password string
	}
	var resultData []resultItem

	tx := db.MySQL.Begin()
	teamName := "" // Log
	for _, item := range inputForm {
		password := randstr.String(16)
		newTeam := &db.Team{
			Name:      item.Name,
			Password:  utils.AddSalt(password),
			Logo:      item.Logo,
			SecretKey: randstr.Hex(16),
		}
		if tx.Create(newTeam).RowsAffected != 1 {
			tx.Rollback()
			return utils.MakeErrJSON(500, 50002,
				locales.I18n.T(c.GetString("lang"), "team.post_error"),
			)
		}
		resultData = append(resultData, resultItem{
			Name:     item.Name,
			Password: password,
		})
		teamName += item.Name + ", "
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.new_team", gin.H{
			"count":    len(inputForm),
			"teamName": teamName,
		})),
	)
	return utils.MakeSuccessJSON(resultData)
}

// EditTeam is edit a team info handler.
func EditTeam(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID   uint   `binding:"required"`
		Name string `binding:"required"`
		Logo string // Logo is not necessary.
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40005,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	// Check the team existed or not.
	var count int
	db.MySQL.Model(db.Team{}).Where(&db.Team{Model: gorm.Model{ID: inputForm.ID}}).Count(&count)
	if count == 0 {
		return utils.MakeErrJSON(404, 40401,
			locales.I18n.T(c.GetString("lang"), "team.not_found"),
		)
	}

	// Check the team name repeated or not.
	var repeatCheckTeam db.Team
	db.MySQL.Model(db.Team{}).Where(&db.Team{Name: inputForm.Name}).Find(&repeatCheckTeam)
	if repeatCheckTeam.Name != "" && repeatCheckTeam.ID != inputForm.ID {
		return utils.MakeErrJSON(400, 40004,
			locales.I18n.T(c.GetString("lang"), "team.repeat"),
		)
	}

	tx := db.MySQL.Begin()
	if tx.Model(&db.Team{}).Where(&db.Team{Model: gorm.Model{ID: inputForm.ID}}).Updates(gin.H{
		"Name": inputForm.Name,
		"Logo": inputForm.Logo,
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50003,
			locales.I18n.T(c.GetString("lang"), "team.put_error"),
		)
	}
	tx.Commit()

	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "team.put_success"))
}

// DeleteTeam is delete a team handler.
func DeleteTeam(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40006,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40006,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	var team db.Team
	db.MySQL.Where(&db.Team{Model: gorm.Model{ID: uint(id)}}).Find(&team)
	if team.Name == "" {
		return utils.MakeErrJSON(404, 40401,
			locales.I18n.T(c.GetString("lang"), "team.not_found"),
		)
	}

	tx := db.MySQL.Begin()
	if tx.Where("id = ?", uint(id)).Delete(&db.Team{}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50004,
			locales.I18n.T(c.GetString("lang"), "team.delete_error"),
		)
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.delete_team", gin.H{
			"teamName": team.Name,
		})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "team.delete_success"))
}

// ResetTeamPassword will reset a team's password. The new password is a random string.
func ResetTeamPassword(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID uint `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40007,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	// Check the team existed or not.
	var checkTeam db.Team
	db.MySQL.Model(db.Team{}).Where(&db.Team{Model: gorm.Model{ID: inputForm.ID}}).Find(&checkTeam)
	if checkTeam.Name == "" {
		return utils.MakeErrJSON(404, 40401,
			locales.I18n.T(c.GetString("lang"), "team.not_found"),
		)
	}

	newPassword := randstr.String(16)
	tx := db.MySQL.Begin()
	if tx.Model(&db.Team{}).Where(&db.Team{Model: gorm.Model{ID: inputForm.ID}}).Updates(&db.Team{Password: utils.AddSalt(newPassword)}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50005,
			locales.I18n.T(c.GetString("lang"), "team.reset_password_error"),
		)
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.team_reset_password", gin.H{
			"teamName": checkTeam.Name,
		})),
	)
	return utils.MakeSuccessJSON(newPassword)
}
