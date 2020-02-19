package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"runtime"
)

// Log levels
const (
	NORMAL    = 0
	WARNING   = 1
	IMPORTANT = 2
)

// Log is a gorm model for database table `logs`.
type Log struct {
	gorm.Model

	Level   int // 0 - Normal, 1 - Warning, 2 - Important
	Kind    string
	Content string
}

// NewLog create a new log record in database.
func (s *Service) NewLog(level int, kind string, content string) {
	s.Mysql.Create(&Log{
		Level:   level,
		Kind:    kind,
		Content: content,
	})
}

// GetLogs returns the latest 30 logs.
func (s *Service) GetLogs(c *gin.Context) (int, interface{}) {
	var logs []Log
	s.Mysql.Model(&Log{}).Order("`id` DESC").Limit(30).Find(&logs)
	return s.makeSuccessJSON(logs)
}

// Panel returns the system runtime status, which is used in backstage data panel.
func (s *Service) Panel(c *gin.Context) (int, interface{}) {
	var submitFlag int
	s.Mysql.Model(&AttackAction{}).Count(&submitFlag)

	var checkDown int
	s.Mysql.Model(&DownAction{}).Count(&checkDown)

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	return s.makeSuccessJSON(gin.H{
		"SubmitFlag":         submitFlag,
		"CheckDown":          checkDown,
		"NumGoroutine":       runtime.NumGoroutine(),          // Goroutine number
		"MemAllocated":       s.FileSize(int64(m.Alloc)),      // Allocated memory
		"MemTotal":           s.FileSize(int64(m.TotalAlloc)), // Total memory usage
		"TotalScore":         s.TotalScore(),
		"PreviousRoundScore": s.PreviousRoundScore(),
	})
}
