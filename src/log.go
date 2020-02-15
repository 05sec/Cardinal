package main

import "github.com/jinzhu/gorm"

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
