package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/vidar-team/Cardinal/src/locales"
	"log"
	"os"
)

var conf *Config

// Init, init the config file.
func Init() {
	if os.Getenv("TRAVIS") != "true" {
		_, err := toml.DecodeFile("./conf/Cardinal.toml", &conf)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println(locales.I18n.T(conf.SystemLanguage, "config.load_success"))
	} else {
		// Travis CI Test, set the config in test code.
		conf = new(Config)
		log.Println("Test mode")
	}
}

// Get returns the config struct.
func Get() *Config {
	return conf
}
