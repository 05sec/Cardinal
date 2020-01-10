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

	s.Router = r
	panic(r.Run(s.Conf.Base.Port))
}
