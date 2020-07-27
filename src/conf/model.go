package conf

import (
	"github.com/jinzhu/gorm"
	"time"
)

// Config is the config of the cardinal.
type Config struct {
	Base  `toml:"base";gorm:"-"`
	MySQL `toml:"mysql";gorm:"-"`

	*DynamicConfig
}

// DynamicConfig is the config which is stored in database.
// So it's a GORM model for users can edit it anytime.
type DynamicConfig struct {
	gorm.Model `json:"-"`

	// TODO
	//Title string `gorm:"default:'HCTF'";json:"Title";binding:"required"`
}

// Base is the basic config of the cardinal.
type Base struct {
	Title            string
	SystemLanguage   string
	BeginTime        time.Time
	RestTime         [][]time.Time
	EndTime          time.Time
	Duration         uint
	SeparateFrontend bool
	Sentry           bool
	Port             string
	Salt             string
	FlagPrefix       string
	FlagSuffix       string
	CheckDownScore   int
	AttackScore      int
}

// MySQL is the database config of the cardinal.
type MySQL struct {
	DBHost     string
	DBUsername string
	DBPassword string
	DBName     string
}
