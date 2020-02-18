package main

import (
	"github.com/BurntSushi/toml"
	"log"
	"time"
)

// Config is the `cardinal.toml` config file struct.
type Config struct {
	Base  `toml:"base"`
	MySQL `toml:"mysql"`
}

// Base is the basic config of the cardinal.
type Base struct {
	Title          string
	BeginTime      time.Time
	RestTime       [][]time.Time
	EndTime        time.Time
	Duration       uint
	Port           string
	Salt           string
	FlagPrefix     string
	FlagSuffix     string
	CheckDownScore int
	AttackScore    int
}

// MySQL is the database config of the cardinal.
type MySQL struct {
	DBHost     string
	DBUsername string
	DBPassword string
	DBName     string
}

// initConfig will decode the config file and put it into `s.Conf`, so we can get the config globally.
func (s *Service) initConfig() {
	var conf *Config
	_, err := toml.DecodeFile("./conf/Cardinal.toml", &conf)
	if err != nil {
		log.Fatalln(err)
	}

	s.Conf = conf
	log.Println("加载配置文件成功")
}
