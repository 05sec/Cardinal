package main

import (
	"github.com/BurntSushi/toml"
	"log"
)

type Config struct {
	MySQL `toml:"mysql"`
	Redis `toml:"redis"`
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
