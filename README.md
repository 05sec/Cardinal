![Cardinal Logo](https://img.cdn.n3ko.co/lsky/2020/02/16/e75b82afd0932.png)

# Cardinal —— CTF AWD 线下赛平台

[![Build](https://travis-ci.com/vidar-team/Cardinal.svg?branch=master)](https://travis-ci.org/vidar-team/Cardinal)
[![GoReport](https://goreportcard.com/badge/github.com/vidar-team/Cardinal)](https://goreportcard.com/report/github.com/vidar-team/Cardinal)
[![codecov](https://codecov.io/gh/vidar-team/Cardinal/branch/master/graph/badge.svg)](https://codecov.io/gh/vidar-team/Cardinal)

## 介绍

Cardinal 是由 Vidar-Team 开发的 AWD 比赛平台，使用 Go 编写。本程序可以作为 CTF 线下比赛平台，亦可用于团队内部 AWD 模拟练习。

![Cardinal Frontend](https://img.cdn.n3ko.co/lsky/2020/03/03/7b6161f88fb94.png)

![Cardinal Backend](https://img.cdn.n3ko.co/lsky/2020/03/03/a7ccd8a8fbd43.png)

## 功能介绍
* 管理员创建题目、分配题目靶机、参赛队伍、生成 Flag、发布公告
    * 支持上传参赛队伍 Logo
    * 题目可设置状态开放、下线，队伍分数同步更新
    * 批量生成 Flag 并导出，方便 Check bot

* 每轮结束后自动结算分数，并更新排行榜
    * 自动对分数计算正确性进行检查
    * 分数计算异常日志提醒
    * 自定义攻击、Checkdown 分数
    * 队伍平分靶机分数
    
* 管理端首页数据总览查看
    * 管理员、系统重要操作日志记录
    * 系统运行状态查看
    
* 选手查看自己的队伍信息，靶机信息，Token，总排行榜，公告
    * 总排行榜靶机状态实时更新

* 前后端分离，前端开源可定制

## 安装

**由于 Cardinal 目前还处于开发中，会不断进行更新迭代。请选择 GitHub Releases 或 master 分支程序 / 代码。
切勿将 dev 分支的代码用于生产环境中。** 

### 编译安装

#### 部署前端
```
git clone https://github.com/vidar-team/Cardinal_frontend.git
git clone https://github.com/vidar-team/Cardinal_manager_frontend.git
```
分别修改两个前端`utils.js`文件中的`baseURL`为后端接口地址。
```
yarn build
```

#### 部署后端
* 编译安装
```
git clone https://github.com/vidar-team/Cardinal.git
cd Cardinal/src
go build -o Cardinal
./Cardinal
```
运行编译后的二进制文件即可。

* Release 安装

GitHub Releases 下载对应架构的压缩包即可。

* Docker 安装

TODO

## 贡献

[提交 Bug](https://github.com/vidar-team/Cardinal/issues/new) | [Fork & Pull Request](https://github.com/vidar-team/Cardinal/fork)

十分欢迎您和我们一起改进 Cardinal，您可以改进现有程序，加入新功能，完善文档，优化代码等。

[![Contributors](http://ergatejs.implements.io/badges/contributors/vidar-team/Cardinal_1280_96_10.png)](https://github.com/vidar-team/Cardinal/graphs/contributors)

## 协议与许可

© Vidar-Team

使用 APACHE LICENSE VERSION 2.0 进行许可，禁止商业用途。
