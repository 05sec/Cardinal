package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/thanhpk/randstr"
	"github.com/vidar-team/Cardinal/src/conf"
	"github.com/vidar-team/Cardinal/src/locales"
	"github.com/vidar-team/Cardinal/src/utils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const configTemplate = `
[base]
Title="{{ .Title }}"
SystemLanguage="zh-CN"
BeginTime="{{ .BeginTime }}"
RestTime=[
#    ["2020-02-16T17:00:00+08:00","2020-02-16T18:00:00+08:00"],
]
EndTime="{{ .EndTime }}"
Duration={{ .Duration }}
SeparateFrontend={{ .SeparateFrontend }}

Salt="{{ .Salt }}"

Port=":{{ .Port }}"

FlagPrefix="{{ .FlagPrefix }}"
FlagSuffix="{{ .FlagSuffix }}"

CheckDownScore={{ .CheckDownScore }}
AttackScore={{ .AttackScore }}

[mysql]
DBHost="{{ .DBHost }}"
DBUsername="{{ .DBUsername }}"
DBPassword="{{ .DBPassword }}"
DBName="{{ .DBName }}"
`

func (s *Service) install() {
	// Check `uploads` folder exist
	if !utils.FileIsExist("./uploads") {
		err := os.Mkdir("./uploads", os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Check `conf` folder exist
	if !utils.FileIsExist("./conf") {
		err := os.Mkdir("./conf", os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Check `locales` folder exist
	if !utils.FileIsExist("./locales") {
		err := os.Mkdir("./locales", os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Check language file exist
	files, err := ioutil.ReadDir("./locales")
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln("Can not find the language file!")
	}

	if !utils.FileIsExist("./conf/Cardinal.toml") {
		log.Println("Please select a preferred language for the installation guide:")
		for index, lang := range languages {
			fmt.Printf("%s - %s\n", index, lang)
		}
		fmt.Println("")

		// Select a language
		index := "0"
		for languages[index] == "" {
			utils.InputString(&index, "type 1, 2... to select")
		}

		content, err := s.GenerateConfigFileGuide(languages[index])
		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile("./conf/Cardinal.toml", content, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(locales.I18n.T(languages[index], "install.create_config_success"))
	}
}

// GenerateConfigFileGuide can lead the user to fill in the config file.
func (s *Service) GenerateConfigFileGuide(lang string) ([]byte, error) {
	input := struct {
		Title, BeginTime, RestTime, EndTime, SeparateFrontend, Duration, Port, Salt, FlagPrefix, FlagSuffix, CheckDownScore, AttackScore, DBHost, DBUsername, DBPassword, DBName string
	}{
		SeparateFrontend: "false",
		Duration:         "2",
		Port:             "19999",
		FlagPrefix:       "hctf{",
		FlagSuffix:       "}",
		CheckDownScore:   "50",
		AttackScore:      "50",
		DBHost:           "localhost:3306",
		DBName:           "cardinal",
	}

	log.Println(locales.I18n.T(lang, "install.greet"))

	utils.InputString(&input.Title, string(locales.I18n.T(lang, "install.input_title")))

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
	utils.InputString(&input.FlagPrefix, string(locales.I18n.T(lang, "install.flag_prefix")))
	utils.InputString(&input.FlagSuffix, string(locales.I18n.T(lang, "install.flag_suffix")))
	utils.InputString(&input.CheckDownScore, string(locales.I18n.T(lang, "install.checkdown_score")))
	utils.InputString(&input.AttackScore, string(locales.I18n.T(lang, "install.attack_score")))
	utils.InputString(&input.SeparateFrontend, string(locales.I18n.T(lang, "install.separate_frontend")))
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

func (s *Service) initManager() {
	var managerCount int
	s.Mysql.Model(&Manager{}).Count(&managerCount)
	if managerCount == 0 {
		// Create manager account if managers table is empty.
		var managerName, managerPassword string
		utils.InputString(&managerName, string(locales.I18n.T(conf.Get().SystemLanguage, "install.manager_name")))
		utils.InputString(&managerPassword, string(locales.I18n.T(conf.Get().SystemLanguage, "install.manager_password")))
		s.Mysql.Create(&Manager{
			Name:     managerName,
			Password: utils.AddSalt(managerPassword),
		})
		s.NewLog(WARNING, "system", string(locales.I18n.T(conf.Get().SystemLanguage, "install.manager_success")))
		log.Println(locales.I18n.T(conf.Get().SystemLanguage, "install.manager_success"))
	}
}
