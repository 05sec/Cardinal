package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/vidar-team/Cardinal/src/conf"
	"github.com/vidar-team/Cardinal/src/locales"
	"log"
)

func (s *Service) initMySQL() {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local&charset=utf8mb4,utf8",
		conf.Get().DBUsername,
		conf.Get().DBPassword,
		conf.Get().DBHost,
		conf.Get().DBName,
	))

	db.DB().SetMaxIdleConns(128)
	db.DB().SetMaxOpenConns(256)

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
		&WebHook{},
	)

	// Test the database charset.
	if s.Mysql.Exec("SELECT * FROM `logs` WHERE `Content` = '中文测试';").Error != nil {
		log.Fatalln(locales.I18n.T(conf.Get().SystemLanguage, "general.database_charset_error"))
	}

}
