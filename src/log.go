package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"runtime"
)

const (
	NORMAL    = 0
	WARNING   = 1
	IMPORTANT = 2
)

type Log struct {
	gorm.Model

	Level   int // 0 - Normal, 1 - Warning, 2 - Important
	Kind    string
	Content string
}

func (s *Service) NewLog(level int, kind string, content string) {
	s.Mysql.Create(&Log{
		Level:   level,
		Kind:    kind,
		Content: content,
	})
}

func (s *Service) GetLogs(c *gin.Context) (int, interface{}) {
	var logs []Log
	s.Mysql.Model(&Log{}).Order("`id` DESC").Limit(30).Find(&logs)
	return s.makeSuccessJSON(logs)
}

func (s *Service) Panel(c *gin.Context) (int, interface{}) {
	var submitFlag int
	s.Mysql.Model(&AttackAction{}).Count(&submitFlag)

	var checkDown int
	s.Mysql.Model(&DownAction{}).Count(&checkDown)

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	return s.makeSuccessJSON(gin.H{
		"SubmitFlag":   submitFlag,                      // 提交 Flag 数
		"CheckDown":    checkDown,                       // Check Down 次数
		"NumGoroutine": runtime.NumGoroutine(),          // Goroutine 数
		"MemAllocated": s.FileSize(int64(m.Alloc)),      // 内存占用量
		"MemTotal":     s.FileSize(int64(m.TotalAlloc)), // 内存使用量
	})
}
