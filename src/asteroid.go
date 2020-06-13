package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/src/asteroid"
	"github.com/vidar-team/Cardinal/src/utils"
)

func (s *Service) asteroidGreetData() (result asteroid.Greet) {
	var asteroidTeam []asteroid.Team
	var teams []Team
	s.Mysql.Model(&Team{}).Order("score DESC").Find(&teams)
	for rank, team := range teams {
		asteroidTeam = append(asteroidTeam, asteroid.Team{
			Id:    int(team.ID),
			Name:  team.Name,
			Rank:  rank + 1,
			Score: int(team.Score),
		})
	}

	result.Team = asteroidTeam
	result.Time = s.Timer.RoundRemainTime
	result.Round = s.Timer.NowRound
	return
}

func (s *Service) getAsteroidStatus() (int, interface{}) {
	return utils.MakeSuccessJSON(gin.H{
		"status": s.GetBool("asteroid_enabled"),
	})
}
