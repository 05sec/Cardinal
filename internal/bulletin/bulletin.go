package bulletin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/utils"
)

// GetAllBulletins returns all bulletins from the database.
func GetAllBulletins(c *gin.Context) (int, interface{}) {
	var bulletins []db.Bulletin
	db.MySQL.Model(&db.Bulletin{}).Order("`id` DESC").Find(&bulletins)
	return utils.MakeSuccessJSON(bulletins)
}

// NewBulletin is post new bulletin handler for manager.
func NewBulletin(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Title   string `binding:"required"`
		Content string `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40031,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	tx := db.MySQL.Begin()
	if tx.Create(&db.Bulletin{
		Title:   inputForm.Title,
		Content: inputForm.Content,
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50019,
			locales.I18n.T(c.GetString("lang"), "bulletin.post_error"),
		)
	}
	tx.Commit()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "bulletin.post_success"))
}

// EditBulletin is edit new bulletin handler for manager.
func EditBulletin(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID      uint   `binding:"required"`
		Title   string `binding:"required"`
		Content string `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40031,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	var checkBulletin db.Bulletin
	db.MySQL.Where(&db.Bulletin{Model: gorm.Model{ID: inputForm.ID}}).Find(&checkBulletin)
	if checkBulletin.ID == 0 {
		return utils.MakeErrJSON(404, 40404,
			locales.I18n.T(c.GetString("lang"), "bulletin.not_found"),
		)
	}

	newBulletin := &db.Bulletin{
		Title:   inputForm.Title,
		Content: inputForm.Content,
	}
	tx := db.MySQL.Begin()
	if tx.Model(&db.Bulletin{}).Where(&db.Bulletin{Model: gorm.Model{ID: inputForm.ID}}).Updates(&newBulletin).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50020,
			locales.I18n.T(c.GetString("lang"), "bulletin.put_error"),
		)
	}
	tx.Commit()

	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "bulletin.put_success"))
}

// DeleteBulletin is delete new bulletin handler for manager.
func DeleteBulletin(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40032,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40032,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	var checkBulletin db.Bulletin
	db.MySQL.Where(&db.Bulletin{Model: gorm.Model{ID: uint(id)}}).Find(&checkBulletin)
	if checkBulletin.ID == 0 {
		return utils.MakeErrJSON(404, 40404,
			locales.I18n.T(c.GetString("lang"), "bulletin.not_found"),
		)
	}

	tx := db.MySQL.Begin()
	if tx.Where("id = ?", id).Delete(&db.Bulletin{}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50021,
			locales.I18n.T(c.GetString("lang"), "bulletin.delete_error"),
		)
	}
	tx.Commit()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "bulletin.delete_success"))
}
