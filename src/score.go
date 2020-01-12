package main

import "github.com/jinzhu/gorm"

type Score struct {
	gorm.Model

	TeamID    uint
	GameBoxID uint
	Round     int
	Reason    string
	Score     float64 `gorm:"index"`
}

func (s *Service) NewRoundCalculateScore() {
	nowRound := s.Timer.NowRound
	lastRound := nowRound - 1

	// 被 CheckDown 减分
	s.MinusCheckDown(lastRound)
	// 未被 CheckDown 加分
	s.AddCheckDown(lastRound)
}

// 被 CheckDown 减分
func (s *Service) MinusCheckDown(round int) {
	// 获取 CheckDown 记录给对应的 GameBox 减分
	var downActions []DownAction
	s.Mysql.Model(&DownAction{}).Where(&DownAction{Round: round}).Find(&downActions)

	for _, action := range downActions {
		s.Mysql.Create(&Score{
			TeamID:    action.TeamID,
			GameBoxID: action.GameBoxID,
			Round:     round,
			Reason:    "checkdown",
			Score:     float64(-s.Conf.CheckDownScore),
		})
	}
}

// 未被 CheckDown 加分
func (s *Service) AddCheckDown(round int) {
	// 遍历 Challenge
	var challenges []Challenge
	s.Mysql.Model(&Challenge{}).Find(&challenges)
	for _, challenge := range challenges {
		// 获取该题被 CheckDown 的队伍
		var downActions []DownAction
		s.Mysql.Model(&DownAction{}).Where(&DownAction{ChallengeID: challenge.ID, Round: round}).Find(&downActions)
		totalScore := len(downActions) * s.Conf.CheckDownScore // 可供平分的分数

		var downGameBoxID []uint // 被攻陷的 GameBox IDs
		for _, action := range downActions {
			downGameBoxID = append(downGameBoxID, action.GameBoxID)
		}

		// 获取该题未被 CheckDown 的队伍（排除法）
		var safeGameBoxes []GameBox
		s.Mysql.Model(&GameBox{}).Where(&GameBox{ChallengeID: challenge.ID}).Not("id", downGameBoxID).Find(&safeGameBoxes)
		score := float64(totalScore) / float64(len(safeGameBoxes))

		// 加分
		for _, gamebox := range safeGameBoxes {
			s.Mysql.Create(&Score{
				TeamID:    gamebox.TeamID,
				GameBoxID: gamebox.ID,
				Round:     round,
				Reason:    "service_online",
				Score:     score,
			})
		}
	}
}
