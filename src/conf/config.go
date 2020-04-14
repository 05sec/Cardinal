package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/vidar-team/Cardinal/src/locales"
	"log"
)

var conf *Config

func init() {
	_, err := toml.DecodeFile("./conf/Cardinal.toml", &conf)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(locales.I18n.T(conf.SystemLanguage, "config.load_success"))
}

// Get returns the config struct.
func Get() *Config {
	return conf
}
