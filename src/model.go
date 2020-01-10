package main

import "github.com/jinzhu/gorm"

type Manager struct {
	gorm.Model

	Name     string
	Password string
	Token    string // 管理员只允许单点登录
}
