package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// 题目
type Challenge struct {
	gorm.Model
	Title string
}

// 队伍靶机
type GameBox struct {
	gorm.Model
	ChallengeID uint8
	TeamID      uint8

	Description string
	Visible     bool
	Score       int64 // 分数可负
	IsDown      bool
	IsAttacked  bool
}

func (s *Service) GetAllChallenges() (int, interface{}) {
	var challenges []Challenge
	s.Mysql.Model(&Challenge{}).Find(&challenges)
	return s.makeSuccessJSON(challenges)
}

func (s *Service) NewChallenge(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Title string	`binding:"required"`
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	newChallenge := &Challenge{
		Title: inputForm.Title,
	}
	var checkChallenge Challenge

	s.Mysql.Where(newChallenge).Find(&checkChallenge)
	if checkChallenge.Title != ""{
		return s.makeErrJSON(403, 40300, "重复添加")
	}

	tx := s.Mysql.Begin()
	if tx.Create(newChallenge).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "添加 Challenge 失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON("添加 Challenge 成功！")
}
