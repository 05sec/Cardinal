package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/thanhpk/randstr"
	"github.com/vidar-team/Cardinal/src/locales"
	"github.com/vidar-team/Cardinal/src/utils"
	"strconv"
)

// Manager is a gorm model for database table `managers`.
type Manager struct {
	gorm.Model

	Name     string
	Password string `json:"-"`
	IsCheck  bool
	Token    string // For single sign-on
}

// ManagerLoginForm is used for binding login form JSON data.
type ManagerLoginForm struct {
	Name     string `json:"Name" binding:"required"`
	Password string `json:"Password" binding:"required"`
}

// ManagerLogin is manager login handler.
func (s *Service) ManagerLogin(c *gin.Context) (int, interface{}) {
	var formData ManagerLoginForm
	err := c.BindJSON(&formData)
	if err != nil {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	var manager Manager
	s.Mysql.Where(&Manager{Name: formData.Name}).Find(&manager)

	// The check account can't login.
	if manager.ID != 0 && manager.Name != "" && utils.CheckPassword(formData.Password, manager.Password) && !manager.IsCheck {
		// Login successfully
		token := utils.GenerateToken()
		tx := s.Mysql.Begin()
		if tx.Model(&Manager{}).Where(&Manager{Name: manager.Name}).Updates(&Manager{Token: token}).RowsAffected != 1 {
			tx.Rollback()
			return utils.MakeErrJSON(500, 50000,
				locales.I18n.T(c.GetString("lang"), "general.server_error"),
			)
		}
		tx.Commit()
		return utils.MakeSuccessJSON(token)
	}
	return utils.MakeErrJSON(403, 40300,
		locales.I18n.T(c.GetString("lang"), "manager.login_error"),
	)
}

// ManagerLogout is the manager logout handler.
func (s *Service) ManagerLogout(c *gin.Context) (int, interface{}) {
	token := c.GetHeader("Authorization")
	tx := s.Mysql.Begin()
	if token != "" {
		if tx.Model(&Manager{}).Where("`token` = ? AND `is_check` = ?", token, false).Update(map[string]interface{}{"token": ""}).RowsAffected != 1 {
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
func (s *Service) GetAllManager() (int, interface{}) {
	var manager []Manager
	s.Mysql.Model(&Manager{}).Find(&manager)
	return utils.MakeSuccessJSON(manager)
}

// NewManager is add a new manager handler.
func (s *Service) NewManager(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		IsCheck  bool   `json:"IsCheck"`
		Name     string `json:"Name" binding:"required"`
		Password string `json:"Password"` // The check account doesn't need the password.
	}
	var formData InputForm
	err := c.BindJSON(&formData)
	if err != nil {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	if !formData.IsCheck && formData.Password == "" {
		return utils.MakeErrJSON(400, 40001,
			locales.I18n.T(c.GetString("lang"), "manager.error_payload"),
		)
	}

	var checkManager Manager
	s.Mysql.Model(&Manager{}).Where(&Manager{Name: formData.Name}).Find(&checkManager)
	if checkManager.ID != 0 {
		return utils.MakeErrJSON(400, 40001,
			locales.I18n.T(c.GetString("lang"), "manager.repeat"),
		)
	}

	manager := Manager{
		Name:     formData.Name,
		IsCheck:  formData.IsCheck,
		Password: utils.AddSalt(formData.Password),
	}
	tx := s.Mysql.Begin()
	if tx.Create(&manager).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50000,
			locales.I18n.T(c.GetString("lang"), "manager.post_error"),
		)
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.new_manager", gin.H{"name": manager.Name})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "manager.post_success"))
}

// RefreshManagerToken can refresh a manager's token.
// For the check down bot also use a manager account in Cardinal, they can't login by themselves.
func (s *Service) RefreshManagerToken(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	tx := s.Mysql.Begin()
	token := utils.GenerateToken()
	if tx.Model(&Manager{}).Where(&Manager{Model: gorm.Model{ID: uint(id)}}).Update(&Manager{
		Token: token,
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50000,
			locales.I18n.T(c.GetString("lang"), "manager.update_token_fail"),
		)
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.manager_token", gin.H{"id": id})),
	)
	return utils.MakeSuccessJSON(token)
}

// ChangeManagerPassword will change a manager's password to a random string.
func (s *Service) ChangeManagerPassword(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	tx := s.Mysql.Begin()
	password := randstr.String(16)
	if tx.Model(&Manager{}).Where(map[string]interface{}{"id": uint(id), "is_check": false}).Update(&Manager{
		Password: utils.AddSalt(password),
	}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50000,
			locales.I18n.T(c.GetString("lang"), "manager.update_password_fail"),
		)
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.manager_password", gin.H{"id": id})),
	)
	return utils.MakeSuccessJSON(password)
}

// DeleteManager is delete manager handler.
func (s *Service) DeleteManager(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	tx := s.Mysql.Begin()
	if tx.Model(&Manager{}).Where("id = ?", id).Delete(&Manager{}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50000,
			locales.I18n.T(c.GetString("lang"), "manager.delete_error"),
		)
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate",
		string(locales.I18n.T(c.GetString("lang"), "log.delete_manager", gin.H{"id": id})),
	)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "manager.delete_success"))
}
