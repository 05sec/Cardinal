package main

import (
	"github.com/gin-gonic/gin"
)

func (s *Service) initRouter() {
	r := gin.Default()

	s.Router = r
	panic(r.Run(s.Conf.Base.Port))
}
