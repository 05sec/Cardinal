![Cardinal Logo](https://img.cdn.n3ko.co/lsky/2020/02/16/e75b82afd0932.png)

<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
<!-- ALL-CONTRIBUTORS-BADGE:END -->
[![Build](https://travis-ci.com/vidar-team/Cardinal.svg?branch=master)](https://travis-ci.org/vidar-team/Cardinal)
[![GoReport](https://goreportcard.com/badge/github.com/vidar-team/Cardinal)](https://goreportcard.com/report/github.com/vidar-team/Cardinal)
[![codecov](https://codecov.io/gh/vidar-team/Cardinal/branch/master/graph/badge.svg)](https://codecov.io/gh/vidar-team/Cardinal)

# Cardinal —— CTF AWD 线下赛平台
## 介绍

Cardinal 是由 Vidar-Team 开发的 AWD 比赛平台，使用 Go 编写。本程序可以作为 CTF 线下比赛平台，亦可用于团队内部 AWD 模拟练习。

![Cardinal Frontend](https://img.cdn.n3ko.co/lsky/2020/03/03/7b6161f88fb94.png)

![Cardinal Backend](https://img.cdn.n3ko.co/lsky/2020/03/03/a7ccd8a8fbd43.png#)

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
### Release 安装

[下载](https://github.com/vidar-team/Cardinal/releases)适用于您目标机器的架构程序，运行即可。

```
# 解压程序包
tar -zxvf Cardinal_VERSION_OS_ARCH.tar.gz

# 赋予执行权限
chmod +x ./Cardinal

# 运行
./Cardinal
```

### 编译安装

克隆代码，编译后运行生成的二进制文件即可。

```
# 克隆代码
git clone https://github.com/vidar-team/Cardinal.git

# 切换目录
cd Cardinal/src

# 编译
go build -o Cardinal

# 赋予执行权限
chmod +x ./Cardinal

# 运行
./Cardinal
```

## 开始使用
默认端口： `19999`
* 选手端 `http://localhost:19999/`
* 后端管理 `http://localhost:19999/manager`

## Contributors ✨

[提交 Bug](https://github.com/vidar-team/Cardinal/issues/new) | [Fork & Pull Request](https://github.com/vidar-team/Cardinal/fork)

十分欢迎您和我们一起改进 Cardinal，您可以改进现有程序，加入新功能，完善文档，优化代码等。

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<!-- markdownlint-enable -->
<!-- prettier-ignore-end -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

## 协议与许可

© Vidar-Team

使用 APACHE LICENSE VERSION 2.0 进行许可，禁止商业用途。
