package main

import (
	"github.com/jinzhu/gorm"
)

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

	// 攻击加分
	s.AddAttack(lastRound)
	// 被攻击减分
	s.MinusAttack(lastRound)

	// 被 CheckDown 减分
	s.MinusCheckDown(lastRound)
	// 未被 CheckDown 加分
	s.AddCheckDown(lastRound)

	// 刷新分数信息到 MySQL / Redis
	s.RefreshScoreData()
}

func (s *Service) RefreshScoreData() {
	// 更新靶机分数
	var gameBoxes []GameBox
	s.Mysql.Model(&GameBox{}).Find(&gameBoxes)
	for _, gameBox := range gameBoxes {
		var score float64
		s.Mysql.Model(&Score{}).Where(&Score{GameBoxID: gameBox.ID}).Select("sum(Score)").Scan(&score)
		var challenge Challenge
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: gameBox.ChallengeID}}).Find(&challenge)
		s.Mysql.Model(&GameBox{}).Update(&Score{Score: float64(challenge.BaseScore) + score})
	}

	// 更新队伍分数（所有公开靶机分数之和）
	var teams []Team
	s.Mysql.Model(&Team{}).Find(&teams)
	for _, team := range teams {
		var score float64
		s.Mysql.Model(&GameBox{}).Where(&GameBox{TeamID: team.ID, Visible: true}).Select("sum(Score)").Scan(&score)
		s.Mysql.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: team.ID}}).Update(&Team{Score: score})
	}

	// TODO: Redis
}

// 攻击加分
func (s *Service) AddAttack(round int) {
	// 遍历 GameBox
	var gameBoxes []GameBox
	s.Mysql.Model(&GameBox{}).Find(&gameBoxes)
	for _, gameBox := range gameBoxes {
		// 查看该 GameBox 本轮是否被攻击
		var attackActions []AttackAction
		s.Mysql.Model(&AttackAction{}).Where(&AttackAction{GameBoxID: gameBox.ID, Round: round}).Find(&attackActions)
		if len(attackActions) != 0 {
			score := float64(s.Conf.AttackScore) / float64(len(attackActions)) // 攻击者平分的分数
			// 加分
			for _, action := range attackActions {
				// 获取攻击者这道题的 GameBoxID
				var attackerGameBoxID uint
				s.Mysql.Model(&GameBox{}).Where(&GameBox{TeamID: action.AttackerTeamID}).Find(&attackerGameBoxID)

				s.Mysql.Create(&Score{
					TeamID:    action.AttackerTeamID,
					GameBoxID: attackerGameBoxID,
					Round:     round,
					Reason:    "attack",
					Score:     score,
				})
			}
		}
	}
}

// 被攻击减分
func (s *Service) MinusAttack(round int) {
	// 获取本轮 AttackAction
	var attackActions []AttackAction
	s.Mysql.Model(&AttackAction{}).Where(&AttackAction{Round: round}).Find(&attackActions)

	for _, action := range attackActions {
		s.Mysql.Create(&Score{
			TeamID:    action.TeamID,
			GameBoxID: action.GameBoxID,
			Round:     round,
			Reason:    "been_attacked",
			Score:     float64(-s.Conf.AttackScore),
		})
	}
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
