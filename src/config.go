package main

import (
	"github.com/BurntSushi/toml"
	"log"
	"time"
)

type Config struct {
	Base  `toml:"base"`
	MySQL `toml:"mysql"`
	Redis `toml:"redis"`
}

type Base struct {
	Title     string
	BeginTime time.Time
	RestTime  [][]time.Time
	EndTime   time.Time
	Duration  uint
	Port      string
	Salt      string
}

type MySQL struct {
	DBHost     string
	DBUsername string
	DBPassword string
	DBName     string
}

type Redis struct {
	DBHost     string
	DBPort     string
	DBPassword string
}

func (s *Service) initConfig() {
	var conf *Config
	_, err := toml.DecodeFile("./conf/Cardinal.toml", &conf)
	if err != nil {
		log.Fatalln(err)
	}

	s.Conf = conf
	log.Println("加载配置文件成功")
}
