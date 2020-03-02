package main

import (
	"fmt"
	"math"
	"strconv"
)

// HealthyCheck will be used to check whether Cardinal runs normally.
func (s *Service) HealthyCheck() {
	var teamCount int
	s.Mysql.Model(&Team{}).Count(&teamCount)

	previousRoundScore := s.PreviousRoundScore()
	if math.Abs(previousRoundScore) != 0 {
		// If the previous round total score is not equal zero, maybe all the teams were checked down.
		if previousRoundScore != float64(-s.Conf.CheckDownScore*teamCount) {
			// Maybe there are some mistakes in previous round score.
			s.NewLog(IMPORTANT, "healthy_check",
				string(s.I18n.T(s.Conf.Base.SystemLanguage, "healthy.previous_round_non_zero_error")),
			)
		}
	}

	totalScore := s.TotalScore()
	if math.Abs(totalScore) != 0 {
		// If sum all the scores but it is not equal zero, maybe all the teams were checked down in some rounds.
		if int(totalScore)%(s.Conf.CheckDownScore*teamCount) != 0 {
			// Maybe there are some mistakes.
			s.NewLog(IMPORTANT, "healthy_check",
				string(s.I18n.T(s.Conf.Base.SystemLanguage, "healthy.total_score_non_zero_error")),
			)
		}
	}
}

// PreviousRoundScore returns the previous round's score count.
func (s *Service) PreviousRoundScore() float64 {
	var score []float64
	// Pay attention if there is no action in the previous round, the SUM(`score`) will be NULL.
	s.Mysql.Model(&Score{}).Where(&Score{Round: s.Timer.NowRound}).Pluck("IFNULL(SUM(`score`), 0)", &score)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", score[0]), 64)
	return value
}

// TotalScore returns all the rounds' score count.
func (s *Service) TotalScore() float64 {
	var score []float64
	// Pay attention in the first round, the SUM(`score`) is NULL.
	s.Mysql.Model(&Score{}).Pluck("IFNULL(SUM(`score`), 0)", &score)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", score[0]), 64)
	return value
}
