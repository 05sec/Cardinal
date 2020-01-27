package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"math"
	"strconv"
)

// 队伍靶机
type GameBox struct {
	gorm.Model
	ChallengeID uint
	TeamID      uint

	IP          string
	Port        string
	Description string
	Visible     bool
	Score       float64 // 分数可负
	IsDown      bool
	IsAttacked  bool
}

func (s *Service) GetSelfGameBoxes(c *gin.Context) (int, interface{}) {
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
	teamID := c.GetInt("teamID")
	s.Mysql.Table("game_boxes").Where(&GameBox{TeamID: uint(teamID), Visible: true}).Order("challenge_id").Find(&gameBoxes)
	for index, gameBox := range gameBoxes {
		var challenge Challenge
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: gameBox.ChallengeID}}).Find(&challenge)
		gameBoxes[index].Title = challenge.Title
	}
	return s.makeSuccessJSON(gameBoxes)
}

func (s *Service) GetGameBoxes(c *gin.Context) (int, interface{}) {
	pageStr := c.Query("page")   // 当前页
	perPageStr := c.Query("per") // 每页数量

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return s.makeErrJSON(400, 40002, "Error Query")
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage <= 0 {
		return s.makeErrJSON(400, 40002, "Error Query")
	}

	var total int
	s.Mysql.Model(&GameBox{}).Count(&total)
	var gameBox []GameBox
	s.Mysql.Model(&GameBox{}).Offset((page - 1) * perPage).Limit(perPage).Find(&gameBox)

	return s.makeSuccessJSON(gin.H{
		"Data":      gameBox,
		"Total":     total,
		"TotalPage": math.Ceil(float64(total / perPage)),
	})
}

func (s *Service) NewGameBoxes(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ChallengeID uint   `binding:"required"`
		TeamID      uint   `binding:"required"`
		IP          string `binding:"required"`
		Port        string `binding:"required"`
		Description string `binding:"required"`
	}
	var inputForm []InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	for _, item := range inputForm {
		var count int

		// 检查 ChallengeID
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: item.ChallengeID}}).Count(&count)
		if count != 1 {
			return s.makeErrJSON(400, 40001, "Challenge 不存在")
		}

		// 检查 TeamID
		s.Mysql.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: item.TeamID}}).Count(&count)
		if count != 1 {
			return s.makeErrJSON(400, 40001, "Team 不存在")
		}

		// 检查是否重复
		s.Mysql.Model(GameBox{}).Where(&GameBox{ChallengeID: item.ChallengeID, TeamID: item.TeamID}).Count(&count)
		if count != 0 {
			return s.makeErrJSON(400, 40001, "存在重复添加数据")
		}
	}

	tx := s.Mysql.Begin()
	for _, item := range inputForm {
		newGameBox := &GameBox{
			ChallengeID: item.ChallengeID,
			TeamID:      item.TeamID,
			IP:          item.IP,
			Port:        item.Port,
			Description: item.Description,
		}
		if tx.Create(newGameBox).RowsAffected != 1 {
			tx.Rollback()
			return s.makeErrJSON(500, 50000, "添加 GameBox 失败！")
		}
	}
	tx.Commit()
	return s.makeSuccessJSON("添加 GameBox 成功！")
}

func (s *Service) EditGameBox(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID uint `binding:"required"`

		IP          string `binding:"required"`
		Port        string `binding:"required"`
		Description string `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	tx := s.Mysql.Begin()
	if tx.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: inputForm.ID}}).Updates(&GameBox{Description: inputForm.Description}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50001, "修改 GameBox 失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON("修改 GameBox 成功！")
}
