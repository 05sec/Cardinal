package conf

import (
	"os"
	"time"

	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/locales"

	"github.com/BurntSushi/toml"
	"github.com/thanhpk/randstr"
)

var conf *config

func Init() {
	if os.Getenv("CARDINAL_TEST") != "true" {
		_, err := toml.DecodeFile("./conf/Cardinal.toml", &conf)
		if err != nil {
			log.Fatal("Failed to decode config file: %v", err)
		}

		log.Trace(string(locales.I18n.T(conf.SystemLanguage, "config.load_success")))
	} else {
		// Test mode, set the config in test code.
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
				DBHost:     os.ExpandEnv("$DBHOST:$DBPORT"),
				DBUsername: os.Getenv("DBUSER"),
				DBPassword: os.Getenv("DBPASSWORD"),
				DBName:     os.Getenv("DBNAME"),
			},
		}
	}
}

// Get returns the config struct.
func Get() *config {
	return conf
}
