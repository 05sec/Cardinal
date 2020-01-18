package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Service) initRouter() {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders: []string{"Authorization", "Content-type", "User-Agent"},
		AllowOrigins: []string{"*"},
	}))

	// 基础信息
	r.GET("/base", func(c *gin.Context) {
		c.JSON(s.makeSuccessJSON(gin.H{
			"Title": s.Conf.Title,
		}))
	})
	r.GET("/time", func(c *gin.Context) {
		c.JSON(s.getTime())
	})

	// 静态资源
	r.Static("/uploads", "./uploads")

	// 用户登录
	r.POST("/login", func(c *gin.Context) {
		c.JSON(s.TeamLogin(c))
	})
	// 用户登出
	r.GET("/logout", func(c *gin.Context) {
		c.JSON(s.TeamLogout(c))
	})

	// 用户
	team := r.Group("/team")
	team.Use(s.TeamAuthRequired())
	{
		team.POST("/flag", func(c *gin.Context) {
			c.JSON(s.SubmitFlag(c))
		})
		team.GET("/info", func(c *gin.Context) {
			c.JSON(s.GetTeamInfo(c))
		})
		team.GET("/gameboxes", func(c *gin.Context) {
			c.JSON(s.GetSelfGameBoxes(c))
		})
		team.GET("/rank", func(c *gin.Context) {
			c.JSON(s.makeSuccessJSON(gin.H{"Title": s.GetRankListTitle(), "Team": s.GetRankList()}))
		})
		team.GET("/bulletins", func(c *gin.Context) {
			c.JSON(s.GetAllBulletins())
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
		manager.GET("/gameboxes", func(c *gin.Context) {
			c.JSON(s.GetGameBoxes(c))
		})
		manager.POST("/gameboxes", func(c *gin.Context) {
			c.JSON(s.NewGameBoxes(c))
		})
		manager.PUT("/gamebox", func(c *gin.Context) {
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

		// Flag
		manager.POST("/flag/generate", func(c *gin.Context) {
			c.JSON(s.GenerateFlag())
		})

		// Check
		manager.POST("/checkDown", func(c *gin.Context) {
			c.JSON(s.CheckDown(c))
		})

		// Bulletin
		manager.GET("/bulletins", func(c *gin.Context) {
			c.JSON(s.GetAllBulletins())
		})
		manager.POST("/bulletin", func(c *gin.Context) {
			c.JSON(s.NewBulletin(c))
		})
		manager.PUT("/bulletin", func(c *gin.Context) {
			c.JSON(s.NewBulletin(c))
		})
		manager.DELETE("/bulletin", func(c *gin.Context) {
			c.JSON(s.DeleteBulletin(c))
		})

		// File
		manager.POST("/uploadPicture", func(c *gin.Context) {
			c.JSON(s.UploadPicture(c))
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
		if tokenData.ID == 0 {
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
		if managerData.ID == 0 {
			c.JSON(s.makeErrJSON(401, 40100, "未授权访问"))
			c.Abort()
			return
		}

		c.Set("managerData", managerData)
		c.Next()
	}
}
