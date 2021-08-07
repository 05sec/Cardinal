package install

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/dbold"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/utils"
)

// DOCKER_ENV is docker environment sign.
const DOCKER_ENV = "CARDINAL_DOCKER"

func prepare() {
	// Check `uploads` folder exist
	if !utils.FileIsExist("./uploads") {
		err := os.Mkdir("./uploads", os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create ./uploads folder: %v", err)
		}
	}

	// Check `conf` folder exist
	if !utils.FileIsExist("./conf") {
		err := os.Mkdir("./conf", os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create ./conf folder: %v", err)
		}
	}

	// Check `locales` folder exist
	if !utils.FileIsExist("./locales") {
		err := os.Mkdir("./locales", os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create ./locales folder: %v", err)
		}
	}
}

func Init() {
	prepare()

	if utils.FileIsExist("./conf/Cardinal.toml") {
		return
	}

	// Check language file exist
	files, err := ioutil.ReadDir("./locales")
	if err != nil {
		log.Fatal("Failed to read ./locales folder: %v", err)
	}

	languages := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yml" {
			languages = append(languages, strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))
		}
	}
	if len(languages) == 0 {
		log.Fatal("Failed to find the language file!")
	}

	prompt := promptui.Select{
		Label: "Please select a preferred language for the installation guide",
		Items: languages,
	}
	_, language, err := prompt.Run()
	if err != nil {
		log.Fatal("Failed to select language: %v", err)
	}

	conf.App.Language = language

	if err := SetConfigFileGuide(language); err != nil {
		log.Fatal("Failed to start config file guide: %v", err)
	}
	log.Info(string(locales.I18n.T(language, "install.create_config_success")))
}

// SetConfigFileGuide can lead the user to fill in the config file.
func SetConfigFileGuide(lang string) error {
	log.Info(string(locales.I18n.T(lang, "install.greet")))

	// Game start time.
	beginTime, err := inputDateTime(string(locales.I18n.T(lang, "install.begin_time")))
	if err != nil {
		return errors.Wrap(err, "input begin time")
	}
	conf.Game.StartAt = toml.LocalDateTimeOf(beginTime)

	// Game end time.
	endTime, err := inputDateTime(string(locales.I18n.T(lang, "install.end_time")))
	if err != nil {
		return errors.Wrap(err, "input end time")
	}
	conf.Game.EndAt = toml.LocalDateTimeOf(endTime)

	// Game duration.
	duration, err := inputInt(string(locales.I18n.T(lang, "install.duration")), 2)
	if err != nil {
		return errors.Wrap(err, "input duration")
	}
	conf.Game.Duration = uint(duration)

	// App HTTP service port.
	port, err := inputInt(string(locales.I18n.T(lang, "install.port")), 19999, func(port int) error {
		if port <= 0 || port >= 65536 {
			return errors.Errorf("wrong tcp port number: %v", port)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "input port")
	}
	conf.App.HTTPAddr = fmt.Sprintf(":%d", port)

	// Game checkdown score.
	checkdownScore, err := inputInt(string(locales.I18n.T(lang, "install.checkdown_score")), 50, func(score int) error {
		if score < 0 {
			return errors.Errorf("wrong check down score: %v", port)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "input check down score")
	}
	conf.Game.CheckDownScore = checkdownScore

	// Game attack score.
	attackScore, err := inputInt(string(locales.I18n.T(lang, "install.attack_score")), 50, func(score int) error {
		if score < 0 {
			return errors.Errorf("wrong attack score: %v", port)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "input attack score")
	}
	conf.Game.AttackScore = attackScore

	separateFrontend, err := inputConfirm(string(locales.I18n.T(lang, "install.separate_frontend")), false)
	if err != nil {
		return errors.Wrap(err, "input separate frontend")
	}
	conf.App.SeparateFrontend = separateFrontend

	enableSentry, err := inputConfirm(string(locales.I18n.T(lang, "install.sentry")), true)
	if err != nil {
		return errors.Wrap(err, "input separate frontend")
	}
	conf.App.EnableSentry = enableSentry

	databaseHost, err := inputString(string(locales.I18n.T(lang, "install.db_host")), "localhost:3306")
	if err != nil {
		return errors.Wrap(err, "input database host")
	}
	conf.Database.Host = databaseHost

	databaseUser, err := inputString(string(locales.I18n.T(lang, "install.db_username")), "root")
	if err != nil {
		return errors.Wrap(err, "input database user name")
	}
	conf.Database.User = databaseUser

	databasePassword, err := inputPassword(string(locales.I18n.T(lang, "install.db_password")))
	if err != nil {
		return errors.Wrap(err, "input database password")
	}
	conf.Database.Password = databasePassword

	databaseName, err := inputString(string(locales.I18n.T(lang, "install.db_name")), "cardinal")
	if err != nil {
		return errors.Wrap(err, "input database name")
	}
	conf.Database.Name = databaseName

	conf.Database.MaxOpenConns = 200
	conf.Database.MaxIdleConns = 150

	// Generate Salt
	conf.App.SecuritySalt = randstr.String(64)

	if err := conf.Save("./conf/Cardinal.toml"); err != nil {
		return errors.Wrap(err, "save config file")
	}
	return nil
}

func InitManager() {
	var managerCount int
	dbold.MySQL.Model(&dbold.Manager{}).Count(&managerCount)
	if managerCount == 0 {
		var managerName, managerPassword string

		// Check if it is built by docker-compose.
		if os.Getenv(DOCKER_ENV) != "" {
			managerName = "admin_" + randstr.Hex(3)
			managerPassword = randstr.String(16)

			// Print out the account info.
			fmt.Println("\n\n=======================================")
			fmt.Printf("Manager Name: %s\n", managerName)
			fmt.Printf("Manager Password: %s\n", managerPassword)
			fmt.Printf("=======================================\n\n\n")
		} else {
			var err error
			managerName, err = inputString(string(locales.I18n.T(conf.App.Language, "install.manager_name")), "admin")
			if err != nil {
				log.Fatal("Failed to get manager name input: %v", err)
			}
			managerPassword, err = inputPassword(string(locales.I18n.T(conf.App.Language, "install.manager_password")))
			if err != nil {
				log.Fatal("Failed to get manager password input: %v", err)
			}
		}

		// Create manager account if managers table is empty.
		dbold.MySQL.Create(&dbold.Manager{
			Name:     managerName,
			Password: utils.AddSalt(managerPassword),
		})
		logger.New(logger.WARNING, "system", string(locales.I18n.T(conf.App.Language, "install.manager_success")))
		log.Info(string(locales.I18n.T(conf.App.Language, "install.manager_success")))
	}
}
