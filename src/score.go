package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

// Score is a gorm model for database table `scores`.
// Every action (checkdown, attacked...) will be created a score record, and the total score will be calculated by SUM(`score`).
type Score struct {
	gorm.Model

	TeamID    uint
	GameBoxID uint
	Round     int
	Reason    string
	Score     float64 `gorm:"index"`
}

// NewRoundCalculateScore will calculate the score of the previous round.
func (s *Service) NewRoundCalculateScore() {
	nowRound := s.Timer.NowRound
	previousRound := nowRound - 1

	startTime := time.Now().UnixNano()

	// + Attacked score
	s.AddAttack(previousRound)
	// - Been attacked score
	s.MinusAttack(previousRound)

	// - Been check down
	s.MinusCheckDown(previousRound)
	// + Service online
	s.AddCheckDown(previousRound)

	// Calculate and update all the gameboxes' score.
	s.CalculateGameBoxScore()
	// Calculate and update all the teams' score.
	s.CalculateTeamScore()

	// Refresh the ranking list table header.
	//s.SetRankListTitle()
	// Refresh the ranking list.
	s.SetRankList()

	endTime := time.Now().UnixNano()
	s.NewLog(WARNING, "system", fmt.Sprintf("第 %d 轮分数结算完成！耗时 %f s。", previousRound, float64(endTime-startTime)/float64(time.Second)))
}

// CalculateGameBoxScore will calculate all the gameboxes' scores according to the data in scores table.
func (s *Service) CalculateGameBoxScore() {
	var gameBoxes []GameBox
	s.Mysql.Model(&GameBox{}).Find(&gameBoxes)
	for _, gameBox := range gameBoxes {
		var sc struct{ Score float64 `gorm:"Column:Score"` }
		s.Mysql.Table("scores").Select("SUM(score) AS Score").Where("`game_box_id` = ?", gameBox.ID).Scan(&sc)

		var challenge Challenge
		s.Mysql.Model(&Challenge{}).Where(&Challenge{Model: gorm.Model{ID: gameBox.ChallengeID}}).Find(&challenge)                                  // Get the gamebox's base score.
		s.Mysql.Model(&GameBox{}).Where(&GameBox{Model: gorm.Model{ID: gameBox.ID}}).Update(&Score{Score: float64(challenge.BaseScore) + sc.Score}) // Update the gamebox's score.
	}
}

// CalculateTeamScore will Calculate all the teams' score. (By sum the team's gameboxes' scores)
func (s *Service) CalculateTeamScore() {
	var teams []Team
	s.Mysql.Model(&Team{}).Find(&teams)
	for _, team := range teams {
		var sc struct{ Score float64 `gorm:"Column:Score"` }
		s.Mysql.Table("game_boxes").Select("SUM(score) AS Score").Where("`team_id` = ? AND `visible` = ?", team.ID, 1).Scan(&sc)
		s.Mysql.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: team.ID}}).Update(&Team{Score: sc.Score})
	}
}

// AddAttack will add scores to the attacker.
func (s *Service) AddAttack(round int) {
	// Traversal all the gameboxes.
	var gameBoxes []GameBox
	s.Mysql.Model(&GameBox{}).Find(&gameBoxes)
	for _, gameBox := range gameBoxes {
		// This gamebox has been attacked or not.
		var attackActions []AttackAction
		s.Mysql.Model(&AttackAction{}).Where(&AttackAction{GameBoxID: gameBox.ID, Round: round}).Find(&attackActions)
		if len(attackActions) != 0 {
			score := float64(s.Conf.AttackScore) / float64(len(attackActions)) // Score which every attacker can get from this gamebox.
			// Add score to the attackers.
			for _, action := range attackActions {
				// Get the attacker's gamebox ID of this challenge.
				var attackerGameBox GameBox
				s.Mysql.Model(&GameBox{}).Where(&GameBox{TeamID: action.AttackerTeamID, ChallengeID: gameBox.ChallengeID}).Find(&attackerGameBox)

				s.Mysql.Create(&Score{
					TeamID:    action.AttackerTeamID,
					GameBoxID: attackerGameBox.ID,
					Round:     round,
					Reason:    "attack",
					Score:     score,
				})
			}
		}
	}
}

// MinusAttack will minus scores from the victim.
func (s *Service) MinusAttack(round int) {
	var attackActions []struct {
		GameBoxID uint		`gorm:"game_box_id"`
		TeamID    uint		`gorm:"team_id"`
	}

	// Every gamebox can only be deducted once in one round.
	s.Mysql.Table("attack_actions").Select("DISTINCT(`game_box_id`) AS game_box_id, team_id").Where(&AttackAction{Round: round}).Scan(&attackActions)

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

// MinusCheckDown will minus scores from the service down gameboxes.
func (s *Service) MinusCheckDown(round int) {
	// Get all the DownAction of this round.
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

// AddCheckDown will add scores to the service online gameboxes.
func (s *Service) AddCheckDown(round int) {
	// Traversal all the challenges.
	var challenges []Challenge
	s.Mysql.Model(&Challenge{}).Find(&challenges)
	for _, challenge := range challenges {
		// Get the check down teams of this challenge.
		var downActions []DownAction
		s.Mysql.Model(&DownAction{}).Where(&DownAction{ChallengeID: challenge.ID, Round: round}).Find(&downActions)
		totalScore := len(downActions) * s.Conf.CheckDownScore // Score which every online team can get from this challenge.

		// Get the service online teams' Gamebox ID of this challenge.
		// For the score will be added separately into their **gameboxes**.
		var downGameBoxID []uint // Firstly, get the down Gamebox ID of this challenge.
		for _, action := range downActions {
			downGameBoxID = append(downGameBoxID, action.GameBoxID)
		}

		// Then, get the service online Gamebox ID. (Process of elimination)
		var safeGameBoxes []GameBox
		s.Mysql.Model(&GameBox{}).Where(&GameBox{ChallengeID: challenge.ID}).Not("id", downGameBoxID).Find(&safeGameBoxes)
		score := float64(totalScore) / float64(len(safeGameBoxes))

		// Well, add score!
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
