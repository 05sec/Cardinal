package main

import (
	"github.com/gin-gonic/gin"
)

func (s *Service) initRouter() {
	r := gin.Default()
	
	// 用户登录
	r.POST("/login", func(c *gin.Context) {
		c.JSON(s.TeamLogin(c))
	})

	// 用户
	team := r.Group("/team")
	team.Use(s.TeamAuthRequired())
	{
		team.GET("/info", func(c *gin.Context) {
			c.JSON(s.GetTeamInfo(c))
		})
	}

	// 管理员登录
	r.POST("/manager/login", func(c *gin.Context) {
		c.JSON(s.ManagerLogin(c))
	})

	// 管理
	manager := r.Group("/manager")
	manager.Use(s.ManagerAuthRequired())
	{
		// Challenge
		manager.GET("/challenges", func(c *gin.Context) {
			c.JSON(s.GetAllChallenges())
		})
		manager.POST("/challenge", func(c *gin.Context) {
			c.JSON(s.NewChallenge(c))
		})
		manager.PUT("/challenge", func(c *gin.Context) {
			c.JSON(s.EditChallenge(c))
		})
		manager.DELETE("/challenge", func(c *gin.Context) {
			c.JSON(s.DeleteChallenge(c))
		})
		manager.POST("/challenge/visible", func(c *gin.Context) {
			c.JSON(s.SetVisible(c))
		})

		// GameBox
		manager.GET("/gameboxes", func(c *gin.Context){
			c.JSON(s.GetGameBoxes(c))
		})
		manager.POST("/gameboxes", func(c *gin.Context) {
			c.JSON(s.NewGameBoxes(c))
		})
		manager.PUT("/gamebox", func(c *gin.Context){
			c.JSON(s.EditGameBox(c))
		})

		// Team
		manager.GET("/teams", func(c *gin.Context) {
			c.JSON(s.GetAllTeams())
		})
		manager.POST("/teams", func(c *gin.Context) {
			c.JSON(s.NewTeams(c))
		})
		manager.PUT("/team", func(c *gin.Context) {
			c.JSON(s.EditTeam(c))
		})
		manager.POST("/team/resetPassword", func(c *gin.Context) {
			c.JSON(s.ResetTeamPassword(c))
		})
	}

	s.Router = r
	panic(r.Run(s.Conf.Base.Port))
}

// 用户鉴权中间件
func (s *Service) TeamAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(s.makeErrJSON(403, 40300, "未授权访问"))
			c.Abort()
			return
		}

		var tokenData Token
		s.Mysql.Where(&Token{Token: token}).Find(&tokenData)
		if tokenData.ID == 0{
			c.JSON(s.makeErrJSON(401, 40100, "未授权访问"))
			c.Abort()
			return
		}

		c.Set("teamID", tokenData.TeamID)
		c.Next()
	}
}

// 管理员鉴权中间件
func (s *Service) ManagerAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(s.makeErrJSON(403, 40300, "未授权访问"))
			c.Abort()
			return
		}

		var managerData Manager
		s.Mysql.Where(&Manager{Token: token}).Find(&managerData)
		if managerData.ID == 0{
			c.JSON(s.makeErrJSON(401, 40100, "未授权访问"))
			c.Abort()
			return
		}

		c.Set("managerData", managerData)
		c.Next()
	}
}
