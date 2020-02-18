package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

// Challenge is a gorm model for database table `challenges`, used to store the challenges like Web1, Pwn1.
type Challenge struct {
	gorm.Model
	Title     string
	BaseScore int
}

// SetVisible is setting challenge visible status handler.
// When a challenge's visible status changed, all the teams' challenge scores and their total scores will be calculated immediately.
// The ranking list will also be updated.
func (s *Service) SetVisible(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID      uint `binding:"required"`
		Visible bool
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

	s.Mysql.Model(&GameBox{}).Where("challenge_id = ?", inputForm.ID).Update(map[string]interface{}{"visible": inputForm.Visible})

	// Calculate all the teams' score. (Only visible challenges)
	s.CalculateTeamScore()
	// Refresh the ranking list table's header.
	s.SetRankListTitle()
	// Refresh the ranking list teams' scores.
	s.SetRankList()

	status := "不"
	if inputForm.Visible {
		status = ""
	}
	s.NewLog(NORMAL, "manager_operate", fmt.Sprintf("设置 Challenge [ %s ] 状态为%s可见", checkChallenge.Title, status))
	return s.makeSuccessJSON("修改 GameBox 可见状态成功")
}

// GetAllChallenges get all challenges from the database.
func (s *Service) GetAllChallenges() (int, interface{}) {
	var challenges []Challenge
	s.Mysql.Model(&Challenge{}).Find(&challenges)
	type resultStruct struct {
		ID        uint
		CreatedAt time.Time
		Title     string
		Visible   bool
		BaseScore int
	}

	var res []resultStruct
	for _, v := range challenges {
		// For the challenge model doesn't have the `visible` field,
		// We can only get the challenge's visible status by one for its gamebox.
		// TODO: Need to find a better way to get the challenge's visible status.
		var gameBox GameBox
		s.Mysql.Where(&GameBox{ChallengeID: v.ID}).Limit(1).Find(&gameBox)

		res = append(res, resultStruct{
			ID:        v.ID,
			CreatedAt: v.CreatedAt,
			Title:     v.Title,
			Visible:   gameBox.Visible,
			BaseScore: v.BaseScore,
		})
	}
	return s.makeSuccessJSON(res)
}

// NewChallenge is new challenge handler for manager.
func (s *Service) NewChallenge(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Title     string `binding:"required"`
		BaseScore int    `binding:"required"`
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	newChallenge := &Challenge{
		Title:     inputForm.Title,
		BaseScore: inputForm.BaseScore,
	}
	var checkChallenge Challenge

	s.Mysql.Model(&Challenge{}).Where(&Challenge{Title: newChallenge.Title}).Find(&checkChallenge)
	if checkChallenge.Title != "" {
		return s.makeErrJSON(403, 40300, "重复添加")
	}

	tx := s.Mysql.Begin()
	if tx.Create(newChallenge).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50000, "添加 Challenge 失败！")
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate", fmt.Sprintf("新的 Challenge [ %s ] 被创建", newChallenge.Title))
	return s.makeSuccessJSON("添加 Challenge 成功！")
}

// EditChallenge is edit challenge handler for manager.
func (s *Service) EditChallenge(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID        uint   `binding:"required"`
		Title     string `binding:"required"`
		BaseScore int    `binding:"required"`
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
		Title:     inputForm.Title,
		BaseScore: inputForm.BaseScore,
	}
	tx := s.Mysql.Begin()
	if tx.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: inputForm.ID}}).Updates(&newChallenge).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50001, "修改 Challenge 失败！")
	}
	tx.Commit()

	// If the challenge's score is updated, we need to calculate the gameboxes' scores and the teams' scores.
	if inputForm.BaseScore != checkChallenge.BaseScore {
		// Calculate all the teams' score. (Only visible challenges)
		s.CalculateTeamScore()
		// Refresh the ranking list table's header.
		s.SetRankListTitle()
		// Refresh the ranking list teams' scores.
		s.SetRankList()
	}

	// If the challenge's title is updated, we just need to update the ranking list table's header.
	if inputForm.Title != checkChallenge.Title {
		s.SetRankListTitle()
	}

	return s.makeSuccessJSON("修改 Challenge 成功！")
}

// DeleteChallenge is delete challenge handler for manager.
func (s *Service) DeleteChallenge(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return s.makeErrJSON(400, 40000, "Error query")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Query must be number")
	}

	var challenge Challenge
	s.Mysql.Where(&Challenge{Model: gorm.Model{ID: uint(id)}}).Find(&challenge)
	if challenge.Title == "" {
		return s.makeErrJSON(404, 40400, "Challenge 不存在")
	}

	tx := s.Mysql.Begin()
	// 同时删除 GameBox
	tx.Where("challenge_id = ?", uint(id)).Delete(&GameBox{})
	if tx.Where("id = ?", uint(id)).Delete(&Challenge{}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50002, "删除 Challenge 失败！")
	}
	tx.Commit()

	s.NewLog(NORMAL, "manager_operate", fmt.Sprintf("Challenge [ %s ] 被删除", challenge.Title))
	return s.makeSuccessJSON("删除 Challenge 成功！")
}
