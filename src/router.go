package main

import (
	"github.com/gin-gonic/gin"
)

func (s *Service) initRouter() {
	r := gin.Default()

	// 管理员登录
	r.POST("/manager/login", func(c *gin.Context) {
		c.JSON(s.ManagerLogin(c))
	})

	// 管理
	authorized := r.Group("/manager")
	authorized.Use(s.AuthRequired())
	{
		// Challenge
		authorized.GET("/challenges", func(c *gin.Context) {
			c.JSON(s.GetAllChallenges())
		})
		authorized.POST("/challenge", func(c *gin.Context) {
			c.JSON(s.NewChallenge(c))
		})
		authorized.PUT("/challenge", func(c *gin.Context) {
			c.JSON(s.EditChallenge(c))
		})
		authorized.DELETE("/challenge", func(c *gin.Context) {
			c.JSON(s.DeleteChallenge(c))
		})
		authorized.POST("/challenge/visible", func(c *gin.Context) {
			c.JSON(s.SetVisible(c))
		})

		// GameBox
		authorized.GET("/gameboxes", func(c *gin.Context){
			c.JSON(s.GetGameBoxes(c))
		})
		authorized.POST("/gameboxes", func(c *gin.Context) {
			c.JSON(s.NewGameBoxes(c))
		})
		authorized.PUT("/gamebox", func(c *gin.Context){
			c.JSON(s.EditGameBox(c))
		})

		// Team
		authorized.GET("/teams", func(c *gin.Context) {
			c.JSON(s.GetAllTeams())
		})
		authorized.POST("/teams", func(c *gin.Context) {
			c.JSON(s.NewTeams(c))
		})
		authorized.PUT("/team", func(c *gin.Context) {
			c.JSON(s.EditTeam(c))
		})
		authorized.POST("/team/resetPassword", func(c *gin.Context) {
			c.JSON(s.ResetTeamPassword(c))
		})
	}

	s.Router = r
	panic(r.Run(s.Conf.Base.Port))
}

// 鉴权中间件
func (s *Service) AuthRequired() gin.HandlerFunc {
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
