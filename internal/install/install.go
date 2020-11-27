package install

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/thanhpk/randstr"
	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/utils"
	log "unknwon.dev/clog/v2"
)

// DOCKER_ENV: docker environment sign.
const DOCKER_ENV = "CARDINAL_DOCKER"

const configTemplate = `
[base]
SystemLanguage="zh-CN"
BeginTime="{{ .BeginTime }}"
RestTime=[
#    ["2020-02-16T17:00:00+08:00","2020-02-16T18:00:00+08:00"],
]
EndTime="{{ .EndTime }}"
Duration={{ .Duration }}
SeparateFrontend={{ .SeparateFrontend }}
Sentry={{ .Sentry }}

Salt="{{ .Salt }}"

Port=":{{ .Port }}"

CheckDownScore={{ .CheckDownScore }}
AttackScore={{ .AttackScore }}

[mysql]
DBHost="{{ .DBHost }}"
DBUsername="{{ .DBUsername }}"
DBPassword="{{ .DBPassword }}"
DBName="{{ .DBName }}"
`

func Init() {
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

	// Check language file exist
	files, err := ioutil.ReadDir("./locales")
	if err != nil {
		log.Fatal("Failed to read ./locales folder: %v", err)
	}
	languages := map[string]string{}
	index := 0 // Use a outside `index` variable instead of the loop `index`. Not all the files is `.yml`.
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yml" {
			index++
			languages[strconv.Itoa(index)] = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		}
	}
	if len(languages) == 0 {
		log.Fatal("Can not find the language file!")
	}

	if !utils.FileIsExist("./conf/Cardinal.toml") {
		log.Info("Please select a preferred language for the installation guide:")
		for index, lang := range languages {
			fmt.Printf("%s - %s\n", index, lang)
		}
		fmt.Println("")

		// Select a language
		index := "0"
		for languages[index] == "" {
			utils.InputString(&index, "type 1, 2... to select")
		}

		content, err := GenerateConfigFileGuide(languages[index])
		if err != nil {
			log.Fatal("Failed to start config file guide: %v", err)
		}
		err = ioutil.WriteFile("./conf/Cardinal.toml", content, 0644)
		if err != nil {
			log.Fatal("Failed to write config file: %v", err)
		}
		log.Info(string(locales.I18n.T(languages[index], "install.create_config_success")))
	}
}

// GenerateConfigFileGuide can lead the user to fill in the config file.
func GenerateConfigFileGuide(lang string) ([]byte, error) {
	input := struct {
		BeginTime, RestTime, EndTime, SeparateFrontend, Sentry, Duration, Port, Salt, CheckDownScore, AttackScore, DBHost, DBUsername, DBPassword, DBName string
	}{
		SeparateFrontend: "false",
		Sentry:           "true",
		Duration:         "2",
		Port:             "19999",
		CheckDownScore:   "50",
		AttackScore:      "50",
		DBHost:           "localhost:3306",
		DBName:           "cardinal",
	}

	log.Info(string(locales.I18n.T(lang, "install.greet")))

	var beginTime time.Time
	err := errors.New("")
	for err != nil {
		utils.InputString(&input.BeginTime, string(locales.I18n.T(lang, "install.begin_time")))
		beginTime, err = time.ParseInLocation("2006-01-02 15:04:05", input.BeginTime, time.Local)
	}
	input.BeginTime = beginTime.Format(time.RFC3339)

	var endTime time.Time
	err = errors.New("")
	for err != nil {
		utils.InputString(&input.EndTime, string(locales.I18n.T(lang, "install.end_time")))
		endTime, err = time.ParseInLocation("2006-01-02 15:04:05 ", input.EndTime, time.Local)
	}
	input.EndTime = endTime.Format(time.RFC3339)

	utils.InputString(&input.Duration, string(locales.I18n.T(lang, "install.duration")))
	utils.InputString(&input.Port, string(locales.I18n.T(lang, "install.port")))
	utils.InputString(&input.CheckDownScore, string(locales.I18n.T(lang, "install.checkdown_score")))
	utils.InputString(&input.AttackScore, string(locales.I18n.T(lang, "install.attack_score")))
	utils.InputString(&input.SeparateFrontend, string(locales.I18n.T(lang, "install.separate_frontend")))
	utils.InputString(&input.Sentry, string(locales.I18n.T(lang, "install.sentry")))
	utils.InputString(&input.DBHost, string(locales.I18n.T(lang, "install.db_host")))
	utils.InputString(&input.DBUsername, string(locales.I18n.T(lang, "install.db_username")))
	utils.InputString(&input.DBPassword, string(locales.I18n.T(lang, "install.db_password")))
	utils.InputString(&input.DBName, string(locales.I18n.T(lang, "install.db_name")))

	// Generate Salt
	input.Salt = randstr.String(64)

	var wr bytes.Buffer
	configTmpl, err := template.New("").Parse(configTemplate)
	if err != nil {
		return nil, err
	}
	err = configTmpl.Execute(&wr, input)
	if err != nil {
		return nil, err
	}
	return wr.Bytes(), nil
}

func InitManager() {
	var managerCount int
	db.MySQL.Model(&db.Manager{}).Count(&managerCount)
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
			utils.InputString(&managerName, string(locales.I18n.T(conf.Get().SystemLanguage, "install.manager_name")))
			utils.InputString(&managerPassword, string(locales.I18n.T(conf.Get().SystemLanguage, "install.manager_password")))
		}

		// Create manager account if managers table is empty.
		db.MySQL.Create(&db.Manager{
			Name:     managerName,
			Password: utils.AddSalt(managerPassword),
		})
		logger.New(logger.WARNING, "system", string(locales.I18n.T(conf.Get().SystemLanguage, "install.manager_success")))
		log.Info(string(locales.I18n.T(conf.Get().SystemLanguage, "install.manager_success")))
	}
}
