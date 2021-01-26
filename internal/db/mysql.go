package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/locales"
)

var MySQL *gorm.DB

func InitMySQL() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&loc=Local&charset=utf8mb4,utf8",
		conf.Get().DBUsername,
		conf.Get().DBPassword,
		conf.Get().DBHost,
		conf.Get().DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to mysql database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database connection pool: %v", err)
	}
	sqlDB.SetMaxIdleConns(128)
	sqlDB.SetMaxOpenConns(256)

	// Create tables.
	err = db.AutoMigrate(
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
	if err != nil {
		log.Fatal("Failed to migrate tables: %v", err)
	}

	MySQL = db

	// Test the database charset.
	if MySQL.Exec("SELECT * FROM `logs` WHERE `Content` = '中文测试';").Error != nil {
		log.Fatal(string(locales.I18n.T(conf.Get().SystemLanguage, "general.database_charset_error")))
	}
}
