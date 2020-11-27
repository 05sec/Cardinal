package game

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/healthy"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
)

// CalculateRoundScore will calculate the score of the given round.
func CalculateRoundScore(round int) {
	startTime := time.Now().UnixNano()

	// + Attacked score
	addAttack(round)
	// - Been attacked score
	minusAttack(round)

	// - Been check down
	minusCheckDown(round)
	// + Service online
	addCheckDown(round)

	// Calculate and update all the gameboxes' score.
	calculateGameBoxScore()
	// Calculate and update all the teams' score.
	calculateTeamScore()

	// Refresh the ranking list table header.
	//s.SetRankListTitle()
	// Refresh the ranking list.
	SetRankList()

	endTime := time.Now().UnixNano()
	logger.New(logger.WARNING, "system", string(
		locales.I18n.T(conf.Get().SystemLanguage, "log.score_success",
			gin.H{
				"round": round,
				"time":  float64(endTime-startTime) / float64(time.Second),
			}),
	))

	// Do healthy check to make sure the score is correct.
	healthy.HealthyCheck()
}

// calculateGameBoxScore will calculate all the gameboxes' scores according to the data in scores table.
func calculateGameBoxScore() {
	var gameBoxes []db.GameBox
	db.MySQL.Model(&db.GameBox{}).Find(&gameBoxes)
	for _, gameBox := range gameBoxes {
		var sc struct {
			Score float64 `gorm:"Column:Score"`
		}
		db.MySQL.Table("scores").Select("SUM(score) AS Score").Where("`game_box_id` = ?", gameBox.ID).Scan(&sc)

		var challenge db.Challenge
		db.MySQL.Model(&db.Challenge{}).Where(&db.Challenge{Model: gorm.Model{ID: gameBox.ChallengeID}}).Find(&challenge)                                     // Get the gamebox's base score.
		db.MySQL.Model(&db.GameBox{}).Where(&db.GameBox{Model: gorm.Model{ID: gameBox.ID}}).Update(&db.Score{Score: float64(challenge.BaseScore) + sc.Score}) // Update the gamebox's score.
	}
}

// calculateTeamScore will Calculate all the teams' score. (By sum the team's gameboxes' scores)
func calculateTeamScore() {
	var teams []db.Team
	db.MySQL.Model(&db.Team{}).Find(&teams)
	for _, t := range teams {
		var sc struct {
			Score float64 `gorm:"Column:Score"`
		}
		db.MySQL.Table("game_boxes").Select("SUM(score) AS Score").Where("`team_id` = ? AND `visible` = ?", t.ID, 1).Scan(&sc)
		db.MySQL.Model(&db.Team{}).Where(&db.Team{Model: gorm.Model{ID: t.ID}}).Update(&db.Team{Score: sc.Score})
	}
}

// addAttack will add scores to the attacker.
func addAttack(round int) {
	// Traversal all the gameboxes.
	var gameBoxes []db.GameBox
	db.MySQL.Model(&db.GameBox{}).Find(&gameBoxes)
	for _, gameBox := range gameBoxes {
		// This gamebox has been attacked or not.
		var attackActions []db.AttackAction
		db.MySQL.Model(&db.AttackAction{}).Where(&db.AttackAction{GameBoxID: gameBox.ID, Round: round}).Find(&attackActions)
		if len(attackActions) != 0 {
			score := float64(conf.Get().AttackScore) / float64(len(attackActions)) // Score which every attacker can get from this gamebox.
			// Add score to the attackers.
			for _, action := range attackActions {
				// Get the attacker's gamebox ID of this challenge.
				var attackerGameBox db.GameBox
				db.MySQL.Model(&db.GameBox{}).Where(&db.GameBox{TeamID: action.AttackerTeamID, ChallengeID: gameBox.ChallengeID}).Find(&attackerGameBox)

				db.MySQL.Create(&db.Score{
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

// minusAttack will minus scores from the victim.
func minusAttack(round int) {
	var attackActions []struct {
		GameBoxID uint `gorm:"game_box_id"`
		TeamID    uint `gorm:"team_id"`
	}

	// Every gamebox can only be deducted once in one round.
	db.MySQL.Table("attack_actions").Select("DISTINCT(`game_box_id`) AS game_box_id, team_id").Where(&db.AttackAction{Round: round}).Scan(&attackActions)

	for _, action := range attackActions {
		db.MySQL.Create(&db.Score{
			TeamID:    action.TeamID,
			GameBoxID: action.GameBoxID,
			Round:     round,
			Reason:    "been_attacked",
			Score:     float64(-conf.Get().AttackScore),
		})
	}
}

// minusCheckDown will minus scores from the service down gameboxes.
func minusCheckDown(round int) {
	// Get all the DownAction of this round.
	var downActions []db.DownAction
	db.MySQL.Model(&db.DownAction{}).Where(&db.DownAction{Round: round}).Find(&downActions)

	for _, action := range downActions {
		db.MySQL.Create(&db.Score{
			TeamID:    action.TeamID,
			GameBoxID: action.GameBoxID,
			Round:     round,
			Reason:    "checkdown",
			Score:     float64(-conf.Get().CheckDownScore),
		})
	}
}

// addCheckDown will add scores to the service online gameboxes.
func addCheckDown(round int) {
	// Traversal all the challenges.
	var challenges []db.Challenge
	db.MySQL.Model(&db.Challenge{}).Find(&challenges)
	for _, challenge := range challenges {
		// Get the check down teams of this challenge.
		var downActions []db.DownAction
		db.MySQL.Model(&db.DownAction{}).Where(&db.DownAction{ChallengeID: challenge.ID, Round: round}).Find(&downActions)
		totalScore := len(downActions) * conf.Get().CheckDownScore // Score which every online team can get from this challenge.

		// Get the service online teams' Gamebox ID of this challenge.
		// For the score will be added separately into their **gameboxes**.
		var downGameBoxID []uint // Firstly, get the down Gamebox ID of this challenge.
		for _, action := range downActions {
			downGameBoxID = append(downGameBoxID, action.GameBoxID)
		}

		// Then, get the service online Gamebox ID. (Process of elimination)
		var safeGameBoxes []db.GameBox
		db.MySQL.Model(&db.GameBox{}).Where(&db.GameBox{ChallengeID: challenge.ID}).Not("id", downGameBoxID).Find(&safeGameBoxes)
		score := float64(totalScore) / float64(len(safeGameBoxes))

		// Well, add score!
		for _, gamebox := range safeGameBoxes {
			db.MySQL.Create(&db.Score{
				TeamID:    gamebox.TeamID,
				GameBoxID: gamebox.ID,
				Round:     round,
				Reason:    "service_online",
				Score:     score,
			})
		}
	}
}
