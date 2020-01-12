package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"time"
)

// 题目
type Challenge struct {
	gorm.Model
	Title string
}

func (s *Service) SetVisible(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID      int
		Visible bool
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	s.Mysql.Model(&GameBox{}).Where("id = ?", inputForm.ID).Update(map[string]interface{}{"visible": inputForm.Visible})
	return s.makeSuccessJSON("修改 GameBox 可见状态成功")
}

func (s *Service) GetAllChallenges() (int, interface{}) {
	var challenges []Challenge
	s.Mysql.Model(&Challenge{}).Find(&challenges)
	type resultStruct struct {
		ID        uint
		CreatedAt time.Time
		Title     string
		Visible   bool
	}

	var res []resultStruct
	for _, v := range challenges {
		// 获取其中一个靶机的状态来得知 Visible
		var gameBox GameBox
		s.Mysql.Where(&GameBox{ChallengeID: v.ID}).Limit(1).Find(&gameBox)

		res = append(res, resultStruct{
			ID:        v.ID,
			CreatedAt: v.CreatedAt,
			Title:     v.Title,
			Visible:   gameBox.Visible,
		})
	}
	return s.makeSuccessJSON(res)
}

func (s *Service) NewChallenge(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Title string `binding:"required"`
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
	if checkChallenge.Title != "" {
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

func (s *Service) EditChallenge(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID    uint   `binding:"required"`
		Title string `binding:"required"`
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	var checkChallenge Challenge
	s.Mysql.Where(&Challenge{Model: gorm.Model{ID: inputForm.ID}}).Find(&checkChallenge)
	if checkChallenge.Title == "" {
		return s.makeErrJSON(404, 40400, "Challenge 不存在")
	}

	newChallenge := &Challenge{
		Title: inputForm.Title,
	}
	tx := s.Mysql.Begin()
	if tx.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: inputForm.ID}}).Updates(&newChallenge).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50001, "修改 Challenge 失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON("修改 Challenge 成功！")
}

func (s *Service) DeleteChallenge(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID uint `binding:"required"`
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	var challenge Challenge
	s.Mysql.Where(&Challenge{Model: gorm.Model{ID: inputForm.ID}}).Find(&challenge)
	if challenge.Title == "" {
		return s.makeErrJSON(404, 40400, "Challenge 不存在")
	}

	tx := s.Mysql.Begin()
	// 同时删除 GameBox
	tx.Where("challenge_id = ?", inputForm.ID).Delete(&GameBox{})
	if tx.Where("id = ?", inputForm.ID).Delete(&Challenge{}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50002, "删除 Challenge 失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON("删除 Challenge 成功！")
}
