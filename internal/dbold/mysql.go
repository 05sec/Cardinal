package dbold

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/locales"
)

var MySQL *gorm.DB

func InitMySQL() {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local&charset=utf8mb4,utf8",
		conf.Database.User,
		conf.Database.Password,
		conf.Database.Host,
		conf.Database.Name,
	))

	if err != nil {
		log.Fatal("Failed to connect to mysql database: %v", err)
	}

	db.DB().SetMaxIdleConns(conf.Database.MaxIdleConns)
	db.DB().SetMaxOpenConns(conf.Database.MaxOpenConns)

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
		log.Fatal(locales.T("general.database_charset_error"))
	}
}
