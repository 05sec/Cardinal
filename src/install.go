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

		content, err := s.GenerateConfigFileGuide()
		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile("./conf/Cardinal.toml", content, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// GenerateConfigFileGuide can lead the user to fill in the config file.
func (s *Service) GenerateConfigFileGuide() ([]byte, error) {
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

	log.Println("Cardinal.toml 配置文件不存在，安装向导将带领您进行配置。")

	InputString(&input.Title, "请输入比赛名称")

	var beginTime time.Time
	err := errors.New("")
	for err != nil {
		InputString(&input.BeginTime, "请输入比赛开始时间（格式 2020-02-17 12:00:00）")
		beginTime, err = time.ParseInLocation("2006-01-02 15:04:05", input.BeginTime, time.Local)
	}
	input.BeginTime = beginTime.Format(time.RFC3339)

	var endTime time.Time
	err = errors.New("")
	for err != nil {
		InputString(&input.EndTime, "请输入比赛结束时间（格式 2020-02-17 12:00:00）")
		endTime, err = time.ParseInLocation("2006-01-02 15:04:05 ", input.EndTime, time.Local)
	}
	input.EndTime = endTime.Format(time.RFC3339)

	InputString(&input.Duration, "请输入每轮长度（单位：分钟，默认值：2）")
	InputString(&input.Port, "请输入后端服务器端口号（默认值：19999）")
	InputString(&input.FlagPrefix, "请输入 Flag 前缀（默认值：hctf{）")
	InputString(&input.FlagSuffix, "请输入 Flag 后缀（默认值：}）")
	InputString(&input.CheckDownScore, "请输入每次 Checkdown 扣分（默认值：50）")
	InputString(&input.AttackScore, "请输入每次攻击得分（默认值：50）")
	InputString(&input.DBHost, "请输入数据库地址（默认值：localhost:3306）")
	InputString(&input.DBUsername, "请输入数据库账号：")
	InputString(&input.DBPassword, "请输入数据库密码：")
	InputString(&input.DBName, "请输入数据库表名（默认值：cardinal）")

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
		InputString(&managerName, "请输入管理员账号：")
		InputString(&managerPassword, "请输入管理员密码：")
		s.Mysql.Create(&Manager{
			Name:     managerName,
			Password: s.addSalt(managerPassword),
		})
		s.NewLog(WARNING, "system", fmt.Sprintf("添加管理员账号成功，请妥善保管您的账号密码信息！"))
		log.Println("添加管理员账号成功")
	}
}
