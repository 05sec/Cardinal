package conf

import (
	"os"
	"time"

	"github.com/vidar-team/Cardinal/internal/locales"
	log "unknwon.dev/clog/v2"

	"github.com/BurntSushi/toml"
	"github.com/thanhpk/randstr"
)

var conf *config

func Init() {
	if os.Getenv("TRAVIS") != "true" {
		_, err := toml.DecodeFile("./conf/Cardinal.toml", &conf)
		if err != nil {
			log.Fatal("Failed to decode config file: %v", err)
		}

		log.Trace(string(locales.I18n.T(conf.SystemLanguage, "config.load_success")))
	} else {
		// Travis CI Test, set the config in test code.
		conf = new(config)
		log.Trace("Test mode")

		conf = &config{
			Base: Base{
				BeginTime:      time.Now(),
				RestTime:       nil,
				EndTime:        time.Now().Add(12 * time.Hour),
				Duration:       10,
				Port:           ":19999",
				Salt:           randstr.String(64),
				CheckDownScore: 10,
				AttackScore:    10,
			},
			MySQL: MySQL{
				DBHost:     "127.0.0.1:3306",
				DBUsername: "root",
				DBPassword: os.Getenv("TEST_DB_PASSWORD"),
				DBName:     os.Getenv("TEST_DB_NAME"),
			},
		}
	}
}

// Get returns the config struct.
func Get() *config {
	return conf
}
