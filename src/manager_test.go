package main

import (
	"github.com/parnurzeal/gorequest"
	"strings"
	"testing"
)

func TestManagerLogin(t *testing.T){
	req := gorequest.New().Post("http://localhost:19999/manager/login")
	// 登录成功
	req.Data = map[string]interface{}{
		"Name": "E99",
		"Password": "123456",
	}
	resp, body, _ := req.End()
	if resp.StatusCode != 200{
		t.Fatalf("/manager/login 状态码错误：%d", resp.StatusCode)
	}
	if !strings.Contains(body, "success"){
		t.Fatalf("/manager/login 登录失败！")
	}
	// 登录失败
	req.Data = map[string]interface{}{
		"Name": "E99",
		"Password": "asdf",
	}
	resp, body, _ = req.End()
	if resp.StatusCode != 403{
		t.Fatalf("/manager/login 状态码错误：%d", resp.StatusCode)
	}
	if !strings.Contains(body, "账号或密码错误"){
		t.Fatalf("/manager/login 登录失败返回信息错误！[%s]", body)
	}
}
