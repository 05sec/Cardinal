package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/game"
	"github.com/vidar-team/Cardinal/internal/misc"
	"github.com/vidar-team/Cardinal/internal/utils"
	"runtime"
)

// Log levels
const (
	NORMAL = iota
	WARNING
	IMPORTANT
)

// Log is a gorm model for database table `logs`.
type Log struct {
	gorm.Model

	Level   int // 0 - Normal, 1 - Warning, 2 - Important
	Kind    string
	Content string
}

// New create a new log record in database.
func New(level int, kind string, content string) {
	db.MySQL.Create(&Log{
		Level:   level,
		Kind:    kind,
		Content: content,
	})
}

// GetLogs returns the latest 30 logs.
func GetLogs(c *gin.Context) (int, interface{}) {
	var logs []Log
	db.MySQL.Model(&Log{}).Order("`id` DESC").Limit(30).Find(&logs)
	return utils.MakeSuccessJSON(logs)
}

// Panel returns the system runtime status, which is used in backstage data panel.
func Panel(c *gin.Context) (int, interface{}) {
	var submitFlag int
	db.MySQL.Model(&game.AttackAction{}).Count(&submitFlag)

	var checkDown int
	db.MySQL.Model(&game.DownAction{}).Count(&checkDown)

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	return utils.MakeSuccessJSON(gin.H{
		"SubmitFlag":         submitFlag,
		"CheckDown":          checkDown,
		"NumGoroutine":       runtime.NumGoroutine(),         // Goroutine number
		"MemAllocated":       utils.FileSize(int64(m.Alloc)), // Allocated memory
		"TotalScore":         misc.TotalScore(),
		"PreviousRoundScore": misc.PreviousRoundScore(),
		"Version":            utils.VERSION,
		"CommitSHA":          utils.COMMIT_SHA,
		"BuildTime":          utils.BUILD_TIME,
	})
}
