package route

import "github.com/gin-gonic/gin"

// __ is the wrap of the Gin handler.
func __(handler func(*gin.Context) (int, interface{})) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(handler(c))
	}
}
