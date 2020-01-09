package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"time"
)

func (s *Service) initMySQL() {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local&charset=utf8mb4,utf8",
		s.Conf.MySQL.DBUsername,
		s.Conf.MySQL.DBPassword,
		s.Conf.MySQL.DBHost,
		s.Conf.MySQL.DBName,
	))
	if err != nil {
		log.Fatalln(err)
	}

	s.Mysql = db

	// 建表
	s.Mysql.AutoMigrate(&BaseConfig{})

	// 初始化配置数据
	var baseConfigCount int
	s.Mysql.Model(&BaseConfig{}).Count(&baseConfigCount)
	if baseConfigCount == 0 {
		tx := s.Mysql.Begin()
		if tx.Create(&BaseConfig{
			Title:     "HCTF",
			BeginTime: time.Date(2020, 1, 9, 8, 00, 00, 0, time.Local),
			EndTime:   time.Date(2020, 1, 10, 18, 00, 00, 0, time.Local),
			Duration:  10, // 分钟
		}).RowsAffected != 1 {
			tx.Rollback()
			log.Fatalln("初始化配置数据失败！")
		}
		tx.Commit()
		log.Println("初始化配置数据成功")
	}

}
