package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/vidar-team/Cardinal/src/locales"
	"golang.org/x/tools/go/ssa/interp/testdata/src/os"
	"log"
)

var conf *Config

func init() {
	if os.Getenv("TRAVIS") != "true" {
		_, err := toml.DecodeFile("./conf/Cardinal.toml", &conf)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println(locales.I18n.T(conf.SystemLanguage, "config.load_success"))
	} else {
		// Travis CI Test
		log.Println("Test mode")
	}
}

// Get returns the config struct.
func Get() *Config {
	return conf
}
