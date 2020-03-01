package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Service) initRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders: []string{"Authorization", "Content-type", "User-Agent"},
		AllowOrigins: []string{"*"},
	}))

	r.Use(s.I18nMiddleware())

	// Cardinal basic info
	r.Any("/", func(c *gin.Context) {
		c.JSON(s.makeSuccessJSON("Cardinal"))
	})

	r.GET("/base", func(c *gin.Context) {
		c.JSON(s.makeSuccessJSON(gin.H{
			"Title": s.Conf.Title,
		}))
	})
	r.GET("/time", func(c *gin.Context) {
		c.JSON(s.getTime())
	})

	// Static files
	r.Static("/uploads", "./uploads")

	// Team login
	r.POST("/login", func(c *gin.Context) {
		c.JSON(s.TeamLogin(c))
	})
	// Team logout
	r.GET("/logout", func(c *gin.Context) {
		c.JSON(s.TeamLogout(c))
	})

	// Submit flag
	r.POST("/flag", func(c *gin.Context) {
		c.JSON(s.SubmitFlag(c))
	})

	// For team
	team := r.Group("/team")
	team.Use(s.TeamAuthRequired())
	{
		team.GET("/info", func(c *gin.Context) {
			c.JSON(s.GetTeamInfo(c))
		})
		team.GET("/gameboxes", func(c *gin.Context) {
			c.JSON(s.GetSelfGameBoxes(c))
		})
		team.GET("/rank", func(c *gin.Context) {
			c.JSON(s.makeSuccessJSON(gin.H{"Title": s.GetRankListTitle(), "Rank": s.GetRankList()}))
		})
		team.GET("/bulletins", func(c *gin.Context) {
			c.JSON(s.GetAllBulletins())
		})
	}

	// Manager login
	r.POST("/manager/login", func(c *gin.Context) {
		c.JSON(s.ManagerLogin(c))
	})

	// Manager logout
	r.GET("/manager/logout", func(c *gin.Context) {
		c.JSON(s.ManagerLogout(c))
	})

	// For manager
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
		manager.DELETE("/team", func(c *gin.Context) {
			c.JSON(s.DeleteTeam(c))
		})
		manager.POST("/team/resetPassword", func(c *gin.Context) {
			c.JSON(s.ResetTeamPassword(c))
		})

		// Manager
		manager.GET("/managers", func(c *gin.Context) {
			c.JSON(s.GetAllManager())
		})
		manager.POST("/manager", func(c *gin.Context) {
			c.JSON(s.NewManager(c))
		})
		manager.GET("/manager/token", func(c *gin.Context) {
			c.JSON(s.RefreshManagerToken(c))
		})
		manager.GET("/manager/changePassword", func(c *gin.Context) {
			c.JSON(s.ChangeManagerPassword(c))
		})
		manager.DELETE("/manager", func(c *gin.Context) {
			c.JSON(s.DeleteManager(c))
		})

		// Flag
		manager.GET("/flags", func(c *gin.Context) {
			c.JSON(s.GetFlags(c))
		})
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
			c.JSON(s.EditBulletin(c))
		})
		manager.DELETE("/bulletin", func(c *gin.Context) {
			c.JSON(s.DeleteBulletin(c))
		})

		// File
		manager.POST("/uploadPicture", func(c *gin.Context) {
			c.JSON(s.UploadPicture(c))
		})

		// Log
		manager.GET("/logs", func(c *gin.Context) {
			c.JSON(s.GetLogs(c))
		})
		manager.GET("/rank", func(c *gin.Context) {
			c.JSON(s.makeSuccessJSON(gin.H{"Title": s.GetRankListTitle(), "Rank": s.GetManagerRankList()}))
		})
		manager.GET("/panel", func(c *gin.Context) {
			c.JSON(s.Panel(c))
		})
	}

	// 404
	r.NoRoute(func(c *gin.Context) {
		c.JSON(s.makeErrJSON(404, 40400, "not found"))
	})

	// 405
	r.NoMethod(func(c *gin.Context) {
		c.JSON(s.makeErrJSON(405, 40500, "method not allowed"))
	})

	return r
}

// TeamAuthRequired is the team permission check middleware.
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

// ManagerAuthRequired is the manager permission check middleware.
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
