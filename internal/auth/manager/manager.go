package manager

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

// ManagerLogin is manager login handler.
func ManagerLogin(c *gin.Context) (int, interface{}) {
	var formData struct {
		Name     string `binding:"required"`
		Password string `binding:"required"`
	}
	err := c.BindJSON(&formData)
	if err != nil {
		return utils.MakeErrJSON(400, 40008,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	var manager db.Manager
	db.MySQL.Where(&db.Manager{Name: formData.Name}).Find(&manager)

	// The check account can't login.
	if manager.ID != 0 && manager.Name != "" && utils.CheckPassword(formData.Password, manager.Password) && !manager.IsCheck {
		// Login successfully
		token := utils.GenerateToken()
		tx := db.MySQL.Begin()
		if tx.Model(&db.Manager{}).Where(&db.Manager{Name: manager.Name}).Updates(&db.Manager{Token: token}).RowsAffected != 1 {
			tx.Rollback()
			return utils.MakeErrJSON(500, 50006,
				locales.I18n.T(c.GetString("lang"), "general.server_error"),
			)
		}
		tx.Commit()
		return utils.MakeSuccessJSON(token)
	}
	return utils.MakeErrJSON(403, 40303,
		locales.I18n.T(c.GetString("lang"), "manager.login_error"),
	)
}

// ManagerLogout is the manager logout handler.
func ManagerLogout(c *gin.Context) (int, interface{}) {
	token := c.GetHeader("Authorization")
	tx := db.MySQL.Begin()
	if token != "" {
		if tx.Model(&db.Manager{}).Where("`token` = ? AND `is_check` = ?", token, false).Update(map[string]interface{}{"token": ""}).RowsAffected != 1 {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}
	return utils.MakeSuccessJSON(
		locales.I18n.T(c.GetString("lang"), "manager.logout_success"),
	)
}

// GetAllManager returns all the manager.
func GetAllManager(c *gin.Context) (int, interface{}) {
	var manager []db.Manager
	db.MySQL.Model(&db.Manager{}).Find(&manager)
	return utils.MakeSuccessJSON(manager)
}

// NewManager is add a new manager handler.
func NewManager(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		IsCheck  bool   `json:"IsCheck"`
		Name     string `json:"Name" binding:"required"`
		Password string `json:"Password"` // The check account doesn't need the password.
	}
	var formData InputForm
	err := c.BindJSON(&formData)
	if err != nil {
		return utils.MakeErrJSON(400, 40009,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	if !formData.IsCheck && formData.Password == "" {
		return utils.MakeErrJSON(400, 40010,
			locales.I18n.T(c.GetString("lang"), "manager.error_payload"),
		)
	}

	var checkManager db.Manager
	db.MySQL.Model(&db.Manager{}).Where(&db.Manager{Name: formData.Name}).Find(&checkManager)
	if checkManager.ID != 0 {
		return utils.MakeErrJSON(400, 40011,
			locales.I18n.T(c.GetString("lang"), "manager.repeat"),
		)
	}

	manager := db.Manager{
		Name:     formData.Name,
		IsCheck:  formData.IsCheck,
		Password: utils.AddSalt(formData.Password),
	}
	tx := db.MySQL.Begin()
	if tx.Create(&manager).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50007,
			locales.I18n.T(c.GetString("lang"), "manager.post_error"),
		)
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.new_manager", gin.H{"name": manager.Name})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "manager.post_success"))
}

// RefreshManagerToken can refresh a manager's token.
// For the check down bot also use a manager account in Cardinal, they can't login by themselves.
func RefreshManagerToken(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40012,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40012,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	tx := db.MySQL.Begin()
	token := utils.GenerateToken()
	if tx.Model(&db.Manager{}).Where(&db.Manager{Model: gorm.Model{ID: uint(id)}}).Update(&db.Manager{
		Token: token,
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50008,
			locales.I18n.T(c.GetString("lang"), "manager.update_token_fail"),
		)
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.manager_token", gin.H{"id": id})),
	)
	return utils.MakeSuccessJSON(token)
}

// ChangeManagerPassword will change a manager's password to a random string.
func ChangeManagerPassword(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40012,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40012,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	tx := db.MySQL.Begin()
	password := randstr.String(16)
	if tx.Model(&db.Manager{}).Where(map[string]interface{}{"id": uint(id), "is_check": false}).Update(&db.Manager{
		Password: utils.AddSalt(password),
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50009,
			locales.I18n.T(c.GetString("lang"), "manager.update_password_fail"),
		)
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.manager_password", gin.H{"id": id})),
	)
	return utils.MakeSuccessJSON(password)
}

// DeleteManager is delete manager handler.
func DeleteManager(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40012,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40012,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	tx := db.MySQL.Begin()
	if tx.Model(&db.Manager{}).Where("id = ?", id).Delete(&db.Manager{}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50010,
			locales.I18n.T(c.GetString("lang"), "manager.delete_error"),
		)
	}
	tx.Commit()

	logger.New(logger.NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.delete_manager", gin.H{"id": id})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "manager.delete_success"))
}
