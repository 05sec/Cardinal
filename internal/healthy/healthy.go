package healthy

import (
	"fmt"
	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/timer"
	"github.com/vidar-team/Cardinal/locales"
	"math"
	"strconv"
)

// HealthyCheck will be used to check whether Cardinal runs normally.
func HealthyCheck() {
	var teamCount int
	db.MySQL.Model(&db.Team{}).Count(&teamCount)

	previousRoundScore := previousRoundScore()
	if math.Abs(previousRoundScore) != 0 {
		// If the previous round total score is not equal zero, maybe all the teams were checked down.
		if previousRoundScore != float64(-conf.Get().CheckDownScore*teamCount) {
			// Maybe there are some mistakes in previous round score.
			logger.New(logger.IMPORTANT, "healthy_check",
				string(locales.I18n.T(conf.Get().SystemLanguage, "healthy.previous_round_non_zero_error")),
			)
		}
	}

	totalScore := totalScore()
	if math.Abs(totalScore) != 0 {
		// If sum all the scores but it is not equal zero, maybe all the teams were checked down in some rounds.
		if int(totalScore)%(conf.Get().CheckDownScore*teamCount) != 0 {
			// Maybes there are some mistakes.
			logger.New(logger.IMPORTANT, "healthy_check",
				string(locales.I18n.T(conf.Get().SystemLanguage, "healthy.total_score_non_zero_error")),
			)
		}
	}
}

// previousRoundScore returns the previous round's score count.
func previousRoundScore() float64 {
	var score []float64
	// Pay attention if there is no action in the previous round, the SUM(`score`) will be NULL.
	db.MySQL.Model(&db.Score{}).Where(&db.Score{Round: timer.Get().NowRound}).Pluck("IFNULL(SUM(`score`), 0)", &score)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", score[0]), 64)
	return value
}

// totalScore returns all the rounds' score count.
func totalScore() float64 {
	var score []float64
	// Pay attention in the first round, the SUM(`score`) is NULL.
	db.MySQL.Model(&db.Score{}).Pluck("IFNULL(SUM(`score`), 0)", &score)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", score[0]), 64)
	return value
}
