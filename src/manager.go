package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/thanhpk/randstr"
	"strconv"
)

type Manager struct {
	gorm.Model

	Name     string
	Password string `json:"-"`
	Token    string // 管理员只允许单点登录
}

type ManagerLoginForm struct {
	Name     string `json:"Name" binding:"required"`
	Password string `json:"Password" binding:"required"`
}

func (s *Service) ManagerLogin(c *gin.Context) (int, interface{}) {
	var formData ManagerLoginForm
	err := c.BindJSON(&formData)
	if err != nil {
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
	} else {
		return s.makeErrJSON(403, 40300, "账号或密码错误！")
	}
}

func (s *Service) GetAllManager() (int, interface{}) {
	var manager []Manager
	s.Mysql.Model(&Manager{}).Find(&manager)
	return s.makeSuccessJSON(manager)
}

func (s *Service) NewManager(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Name     string `json:"Name" binding:"required"`
		Password string `json:"Password" binding:"required"`
	}
	var formData InputForm
	err := c.BindJSON(&formData)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	var checkManager Manager
	s.Mysql.Model(&Manager{}).Where(&Manager{Name: formData.Name}).Find(&checkManager)
	if checkManager.ID != 0 {
		return s.makeErrJSON(400, 40001, "管理员名称重复")
	}

	manager := Manager{
		Name:     formData.Name,
		Password: s.addSalt(formData.Password),
	}
	tx := s.Mysql.Begin()
	if tx.Create(&manager).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "添加管理员失败！")
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate", fmt.Sprintf("新的管理员账号 [ %s ] 被添加", manager.Name))
	return s.makeSuccessJSON("添加管理员成功！")
}

func (s *Service) RefreshManagerToken(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return s.makeErrJSON(400, 40000, "Error query")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Query must be number")
	}

	tx := s.Mysql.Begin()
	token := s.generateToken()
	if tx.Model(&Manager{}).Where(&Manager{Model: gorm.Model{ID: uint(id)}}).Update(&Manager{
		Token: token,
	}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "更新管理员 Token 失败！")
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate", fmt.Sprintf("管理员 [ ID: %d ] Token 已刷新", id))
	return s.makeSuccessJSON(token)
}

func (s *Service) ChangeManagerPassword(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return s.makeErrJSON(400, 40000, "Error query")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Query must be number")
	}

	tx := s.Mysql.Begin()
	password := randstr.String(16)
	if tx.Model(&Manager{}).Where(&Manager{Model: gorm.Model{ID: uint(id)}}).Update(&Manager{
		Password: s.addSalt(password),
	}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "修改管理员密码失败！")
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate", fmt.Sprintf("管理员 [ ID: %d ] 密码已修改", id))
	return s.makeSuccessJSON(password)
}

func (s *Service) DeleteManager(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return s.makeErrJSON(400, 40000, "Error query")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Query must be number")
	}

	tx := s.Mysql.Begin()
	if tx.Model(&Manager{}).Where("id = ?", id).Delete(&Manager{}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "删除管理员失败")
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate", fmt.Sprintf("管理员 [ ID: %d ] 已删除", id))
	return s.makeSuccessJSON("删除管理员成功")
}
