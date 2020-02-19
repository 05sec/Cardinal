package main

import (
	"fmt"
	"strconv"
)

// Used to check whether Cardinal runs normally.

func (s *Service) HealthyCheck() {

}

// PreviousRoundScoreCount returns the previous round's score count.
func (s *Service) PreviousRoundScoreCount() float64 {
	var score []float64
	s.Mysql.Debug().Model(&Score{}).Where(&Score{Round: s.Timer.NowRound}).Pluck("SUM(`score`)", &score)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", score[0]), 64)
	return value
}

// TotalScoreCount returns all the rounds' score count.
func (s *Service) TotalScoreCount() float64 {
	var score []float64
	s.Mysql.Model(&Score{}).Pluck("SUM(`score`)", &score)
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", score[0]), 64)
	return value
}
