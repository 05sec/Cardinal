package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/thanhpk/randstr"
)

type Team struct {
	gorm.Model

	Name     string
	Password string
	Logo     string
	Score    int64
}

func (s *Service) NewTeams(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Name string `binding:"required"`
		Logo string `binding:"required"`
	}
	var inputForm []InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// 检查是否重复
	for _, item := range inputForm {
		var count int
		s.Mysql.Model(Team{}).Where(&Team{Name: item.Name}).Count(&count)
		if count != 0 {
			return s.makeErrJSON(400, 40001, "存在重复添加数据")
		}
	}

	type resultItem struct {
		Name     string
		Password string
	}
	var resultData []resultItem

	tx := s.Mysql.Begin()
	for _, item := range inputForm {
		password := randstr.String(16)
		newTeam := &Team{
			Name:     item.Name,
			Password: s.addSalt(password),
			Logo:     item.Logo,
		}
		if tx.Create(newTeam).RowsAffected != 1 {
			tx.Rollback()
			return s.makeErrJSON(500, 50000, "添加 Team 失败！")
		}
		resultData = append(resultData, resultItem{
			Name:     item.Name,
			Password: password,
		})
	}
	tx.Commit()
	return s.makeSuccessJSON(resultData)
}
