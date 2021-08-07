package healthy

import (
	"fmt"
	"math"
	"strconv"

	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/dbold"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/timer"
)

// HealthyCheck will be used to check whether Cardinal runs normally.
func HealthyCheck() {
	var teamCount int
	dbold.MySQL.Model(&dbold.Team{}).Count(&teamCount)

	previousRoundScore := PreviousRoundScore()
	if math.Abs(previousRoundScore) != 0 {
		// If the previous round total score is not equal zero, maybe all the teams were checked down.
		if previousRoundScore != float64(-conf.Game.CheckDownScore*teamCount) {
			// Maybe there are some mistakes in previous round score.
			logger.New(logger.IMPORTANT, "healthy_check",
				string(locales.I18n.T(conf.App.Language, "healthy.previous_round_non_zero_error")),
			)
		}
	}

	totalScore := TotalScore()
	if math.Abs(totalScore) != 0 {
		// If sum all the scores but it is not equal zero, maybe all the teams were checked down in some rounds.
		if int(totalScore)%(conf.Game.CheckDownScore*teamCount) != 0 {
			// Maybes there are some mistakes.
			logger.New(logger.IMPORTANT, "healthy_check",
				string(locales.I18n.T(conf.App.Language, "healthy.total_score_non_zero_error")),
			)
		}
	}
}

// PreviousRoundScore returns the previous round's score count.
func PreviousRoundScore() float64 {
	var score []float64
	// Pay attention if there is no action in the previous round, the SUM(`score`) will be NULL.
	dbold.MySQL.Model(&dbold.Score{}).Where(&dbold.Score{Round: timer.Get().NowRound}).Pluck("IFNULL(SUM(`score`), 0)", &score)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", score[0]), 64)
	return value
}

// TotalScore returns all the rounds' score count.
func TotalScore() float64 {
	var score []float64
	// Pay attention in the first round, the SUM(`score`) is NULL.
	dbold.MySQL.Model(&dbold.Score{}).Pluck("IFNULL(SUM(`score`), 0)", &score)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", score[0]), 64)
	return value
}
