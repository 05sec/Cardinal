package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/locales"
	log "unknwon.dev/clog/v2"
)

var MySQL *gorm.DB

func InitMySQL() {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local&charset=utf8mb4,utf8",
		conf.Get().DBUsername,
		conf.Get().DBPassword,
		conf.Get().DBHost,
		conf.Get().DBName,
	))

	if err != nil {
		log.Fatal("Failed to connect to mysql database: %v", err)
	}

	db.DB().SetMaxIdleConns(128)
	db.DB().SetMaxOpenConns(256)

	// Create tables.
	db.AutoMigrate(
		&Manager{},
		&Challenge{},
		&Token{},
		&Team{},
		&Bulletin{},
		&BulletinRead{}, // Not used

		&AttackAction{},
		&DownAction{},
		&Score{},
		&Flag{},
		&GameBox{},

		&Log{},
		&WebHook{},

		&DynamicConfig{},
	)

	MySQL = db

	// Test the database charset.
	if MySQL.Exec("SELECT * FROM `logs` WHERE `Content` = '中文测试';").Error != nil {
		log.Fatal(string(locales.I18n.T(conf.Get().SystemLanguage, "general.database_charset_error")))
	}
}
