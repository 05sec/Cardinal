package main

import (
	"github.com/gin-gonic/gin"
)

type ManagerLoginForm struct {
	Name     string `json:"Name"`
	Password string `json:"Password"`
}

func (s *Service) ManagerLogin(c *gin.Context) (int, interface{}) {
	var formData ManagerLoginForm
	err := c.BindJSON(&formData)
	if err != nil{
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	var manager Manager
	s.Mysql.Where(&Manager{Name: formData.Name}).Find(&manager)

	if manager.Name != "" && s.checkPassword(formData.Password, manager.Password) {
		// 登录成功
		token := s.generateToken()
		tx := s.Mysql.Begin()
		if tx.Model(&Manager{}).Where(&Manager{Name: manager.Name}).Updates(&Manager{Token: token}).RowsAffected != 1 {
			tx.Rollback()
			return s.makeErrJSON(500, 50000, "Server error")
		}
		tx.Commit()
		return s.makeSuccessJSON(token)
	}else{
		return s.makeErrJSON(403, 40300, "账号或密码错误！")
	}
}
