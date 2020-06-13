package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/src/asteroid"
	"github.com/vidar-team/Cardinal/src/locales"
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

func (s *Service) asteroidAttack(c *gin.Context) (int, interface{}) {
	var attackData struct {
		From int `binding:"required"`
		To   int `binding:"required"`
	}
	if err := c.BindJSON(&attackData); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	asteroid.Attack(attackData.From, attackData.To)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func (s *Service) asteroidRank(c *gin.Context) (int, interface{}) {
	asteroid.Rank()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func (s *Service) asteroidStatus(c *gin.Context) (int, interface{}) {
	var status struct {
		Id     int    `binding:"required"`
		Status string `binding:"required"`
	}
	if err := c.BindJSON(&status); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	if status.Status != "down" && status.Status != "attacked" {
		return utils.MakeErrJSON(400, 40039, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	asteroid.Status(status.Id, status.Status)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func (s *Service) asteroidRound(c *gin.Context) (int, interface{}) {
	var round struct {
		Round int `binding:"required"`
	}
	if err := c.BindJSON(&round); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	asteroid.Round(round.Round)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func (s *Service) asteroidEasterEgg(c *gin.Context) (int, interface{}) {
	asteroid.EasterEgg()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func (s *Service) asteroidTime(c *gin.Context) (int, interface{}) {
	var time struct {
		Time int `binding:"required"`
	}
	if err := c.BindJSON(&time); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	asteroid.Time(time.Time)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func (s *Service) asteroidClear(c *gin.Context) (int, interface{}) {
	var clear struct {
		Id int `binding:"required"`
	}
	if err := c.BindJSON(&clear); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	asteroid.Clear(clear.Id)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func (s *Service) asteroidClearAll(c *gin.Context) (int, interface{}) {
	asteroid.ClearAll()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}
