package conf

import (
	"time"
)

type config struct {
	Base  `toml:"base"`
	MySQL `toml:"mysql"`
}

// Base is the basic config of the cardinal.
type Base struct {
	SystemLanguage   string
	BeginTime        time.Time
	RestTime         [][]time.Time
	EndTime          time.Time
	Duration         uint
	SeparateFrontend bool
	Sentry           bool
	Port             string
	Salt             string
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
