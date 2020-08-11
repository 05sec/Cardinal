package route

import "github.com/gin-gonic/gin"

//func __(code int, obj interface{}) func(*gin.Context) {
//	return func(c *gin.Context) {
//		c.JSON(code, obj)
//	}
//}

func __(handler func(*gin.Context) (int, interface{})) func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(handler(c))
	}
}
