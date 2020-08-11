package game

import (
	"github.com/vidar-team/Cardinal/internal/asteroid"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/dynamic_config"
	"github.com/vidar-team/Cardinal/internal/timer"
	"github.com/vidar-team/Cardinal/internal/utils"
)

func AsteroidGreetData() (result asteroid.Greet) {
	var asteroidTeam []asteroid.Team
	var teams []db.Team
	db.MySQL.Model(&db.Team{}).Order("score DESC").Find(&teams)
	for rank, team := range teams {
		asteroidTeam = append(asteroidTeam, asteroid.Team{
			Id:    int(team.ID),
			Name:  team.Name,
			Rank:  rank + 1,
			Image: team.Logo,
			Score: int(team.Score),
		})
	}

	result.Title = dynamic_config.Get(utils.TITLE_CONF)
	result.Team = asteroidTeam
	result.Time = timer.Get().RoundRemainTime
	result.Round = timer.Get().NowRound
	return
}
