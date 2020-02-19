package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
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

	// Create tables.
	s.Mysql.AutoMigrate(
		&Manager{},
		&Challenge{},
		&Token{},
		&Team{},
		&GameBox{},
		&Bulletin{},
		&BulletinRead{},

		&AttackAction{},
		&DownAction{},
		&Score{},
		&Flag{},
		&Log{},
	)

	// Create init data.
	var managerCount int
	s.Mysql.Model(&Manager{}).Count(&managerCount)
	if managerCount == 0 {
		// Create manager account if managers table is empty.
		var managerName, managerPassword string
		InputString(&managerName, "请输入管理员账号：")
		InputString(&managerPassword, "请输入管理员密码：")
		s.Mysql.Create(&Manager{
			Name:     managerName,
			Password: s.addSalt(managerPassword),
		})
		s.NewLog(WARNING, "system", fmt.Sprintf("添加管理员账号成功，请妥善保管您的账号密码信息！"))
		log.Println("添加管理员账号成功")
	}
}
