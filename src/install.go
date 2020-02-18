package main

import (
	"bytes"
	"github.com/thanhpk/randstr"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

const configTemplate = `
[base]
Title="{{ .Title }}"    # 比赛名称
BeginTime="{{ .BeginTime }}"   # 比赛开始时间
RestTime=[
#    ["2020-02-16T17:00:00+08:00","2020-02-16T18:00:00+08:00"],      # 中途暂停区间
]
EndTime="{{ .EndTime }}"      # 比赛结束时间
Duration={{ .Duration }}    # 每轮长度（分钟）

Salt="{{ .Salt }}"    # 务必改成一个谁也猜不到的随机字符串！！

Port=":{{ .Port }}"       # 平台后端服务端口

FlagPrefix="{{ .FlagPrefix }}"  # Flag 前缀
FlagSuffix="{{ .FlagSuffix }}"      # Flag 后缀

CheckDownScore={{ .CheckDownScore }}   # 每次 CheckDown 扣分
AttackScore={{ .AttackScore }}      # 每次攻击得分

[mysql]     # 数据库配置信息
DBHost="{{ .DBHost }}" # 数据库地址
DBUsername="{{ .DBUsername }}"       # 数据库账号
DBPassword="{{ .DBPassword }}"       # 数据库密码
DBName="{{ .DBName }}"       # 数据库表名
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

	if !IsExist("./conf/Cardinal.toml") {
		content, err := s.GenerateConfigFileGuide()
		if err != nil {
			log.Fatalln(err)
		}
		ioutil.WriteFile("./conf/Cardinal.toml", content, 0644)
		log.Println("创建 Cardinal.toml 配置文件成功！")
	}
}

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
	InputString(&input.BeginTime, "请输入比赛开始时间（格式 2020-02-17T12:00:00+08:00）")
	InputString(&input.EndTime, "请输入比赛结束时间（格式 2020-02-17T12:00:00+08:00）")
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
	if err != nil{
		return nil, err
	}
	err = configTmpl.Execute(&wr, input)
	if err != nil {
		return nil, err
	}
	return wr.Bytes(), nil
}
