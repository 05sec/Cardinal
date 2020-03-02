package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/thanhpk/randstr"
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
	if !IsExist("./uploads") {
		err := os.Mkdir("./uploads", os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Check `conf` folder exist
	if !IsExist("./conf") {
		err := os.Mkdir("./conf", os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Check `locales` folder exist
	if !IsExist("./locales") {
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
	for index, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yml" {
			languages[strconv.Itoa(index+1)] = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		}
	}
	if len(languages) == 0 {
		log.Fatalln("Can not find the language file!")
	}

	if !IsExist("./conf/Cardinal.toml") {
		log.Println("Please select a preferred language for the installation guide:")
		for index, lang := range languages {
			fmt.Printf("%s - %s\n", index, lang)
		}
		fmt.Println("")

		// Select a language
		index := "0"
		var err = errors.New("")
		for languages[index] == "" {
			InputString(&index, "type 1, 2... to select")
		}

		content, err := s.GenerateConfigFileGuide(languages[index])
		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile("./conf/Cardinal.toml", content, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(s.I18n.T(languages[index], "install.create_config_success"))
	}
}

// GenerateConfigFileGuide can lead the user to fill in the config file.
func (s *Service) GenerateConfigFileGuide(lang string) ([]byte, error) {
	input := struct {
		Title, BeginTime, RestTime, EndTime, Duration, Port, Salt, FlagPrefix, FlagSuffix, CheckDownScore, AttackScore, DBHost, DBUsername, DBPassword, DBName string
	}{
		Duration:       "2",
		Port:           "19999",
		FlagPrefix:     "hctf{",
		FlagSuffix:     "}",
		CheckDownScore: "50",
		AttackScore:    "50",
		DBHost:         "localhost:3306",
		DBName:         "cardinal",
	}

	log.Println(s.I18n.T(lang, "install.greet"))

	InputString(&input.Title, string(s.I18n.T(lang, "install.input_title")))

	var beginTime time.Time
	err := errors.New("")
	for err != nil {
		InputString(&input.BeginTime, string(s.I18n.T(lang, "install.begin_time")))
		beginTime, err = time.ParseInLocation("2006-01-02 15:04:05", input.BeginTime, time.Local)
	}
	input.BeginTime = beginTime.Format(time.RFC3339)

	var endTime time.Time
	err = errors.New("")
	for err != nil {
		InputString(&input.EndTime, string(s.I18n.T(lang, "install.end_time")))
		endTime, err = time.ParseInLocation("2006-01-02 15:04:05 ", input.EndTime, time.Local)
	}
	input.EndTime = endTime.Format(time.RFC3339)

	InputString(&input.Duration, string(s.I18n.T(lang, "install.duration")))
	InputString(&input.Port, string(s.I18n.T(lang, "install.port")))
	InputString(&input.FlagPrefix, string(s.I18n.T(lang, "install.flag_prefix")))
	InputString(&input.FlagSuffix, string(s.I18n.T(lang, "install.flag_suffix")))
	InputString(&input.CheckDownScore, string(s.I18n.T(lang, "install.checkdown_score")))
	InputString(&input.AttackScore, string(s.I18n.T(lang, "install.attack_score")))
	InputString(&input.DBHost, string(s.I18n.T(lang, "install.db_host")))
	InputString(&input.DBUsername, string(s.I18n.T(lang, "install.db_username")))
	InputString(&input.DBPassword, string(s.I18n.T(lang, "install.db_password")))
	InputString(&input.DBName, string(s.I18n.T(lang, "install.db_name")))

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
		InputString(&managerName, string(s.I18n.T(s.Conf.Base.SystemLanguage, "install.manager_name")))
		InputString(&managerPassword, string(s.I18n.T(s.Conf.Base.SystemLanguage, "install.manager_password")))
		s.Mysql.Create(&Manager{
			Name:     managerName,
			Password: s.addSalt(managerPassword),
		})
		s.NewLog(WARNING, "system", string(s.I18n.T(s.Conf.Base.SystemLanguage, "install.manager_success")))
		log.Println(s.I18n.T(s.Conf.Base.SystemLanguage, "install.manager_success"))
	}
}
