package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/utils"
)

// Log levels
const (
	NORMAL = iota
	WARNING
	IMPORTANT
)

// New create a new log record in database.
func New(level int, kind string, content string) {
	db.MySQL.Create(&db.Log{
		Level:   level,
		Kind:    kind,
		Content: content,
	})
}

// GetLogs returns the latest 30 logs.
func GetLogs(c *gin.Context) (int, interface{}) {
	var logs []db.Log
	db.MySQL.Model(&db.Log{}).Order("`id` DESC").Limit(30).Find(&logs)
	return utils.MakeSuccessJSON(logs)
}
