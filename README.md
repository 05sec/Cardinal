[![Cardinal Logo](https://img.cdn.n3ko.co/lsky/2020/02/16/e75b82afd0932.png)](https://cardinal.ink)

[![Go](https://github.com/vidar-team/Cardinal/actions/workflows/go.yml/badge.svg)](https://github.com/vidar-team/Cardinal/actions/workflows/go.yml)
[![Code Scanning - Action](https://github.com/vidar-team/Cardinal/actions/workflows/codeql.yml/badge.svg)](https://github.com/vidar-team/Cardinal/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/vidar-team/Cardinal/branch/master/graph/badge.svg?token=FZ9WKD0YE4)](https://codecov.io/gh/vidar-team/Cardinal)
[![GoReport](https://goreportcard.com/badge/github.com/vidar-team/Cardinal)](https://goreportcard.com/report/github.com/vidar-team/Cardinal)
[![Crowdin](https://badges.crowdin.net/cardinal/localized.svg)](https://crowdin.com/project/cardinal)
[![Sourcegraph](https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?logo=sourcegraph)](https://sourcegraph.com/github.com/vidar-team/Cardinal)
[![QQ Group](https://img.shields.io/badge/QQ%E7%BE%A4-130818749-blue.svg?logo=Tencent%20QQ)](https://shang.qq.com/wpa/qunwpa?idkey=c6a35c5fbec05fdcd2d2605e08b4b5f8d6e5854471fefd8c03d370d14870b818)

# [Cardinal](https://cardinal.ink) —— CTF AWD 线下赛平台

## 介绍

Cardinal 是由 Vidar-Team 开发的 AWD 比赛平台，使用 Go 编写。本程序可以作为 CTF 线下比赛平台，亦可用于团队内部 AWD 模拟练习。

![Cardinal Frontend](https://s1.ax1x.com/2020/05/28/tVPltI.png)

<details>
<summary>更多图片</summary>

![Cardinal Backend](https://s1.ax1x.com/2020/05/28/tVP1ht.png)

![Asteroid](https://s1.ax1x.com/2020/05/28/tVP6jU.png)
（该 AWD 实时 3D 攻击大屏为项目 [Asteroid](https://github.com/wuhan005/Asteroid)，已适配接入 Cardinal）

</details>

## 文档

### 官方文档  [cardinal.ink](https://cardinal.ink)

> 请您在使用前认真阅读官方使用文档，谢谢 ♪(･ω･)ﾉ

### 教程

[AWD平台搭建–Cardinal](https://cloud.tencent.com/developer/article/1744139)

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
    * 自动更新靶机 Flag
    * 触发 WebHook，接入第三方应用

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

# 编译
go build -o Cardinal

# 赋予执行权限
chmod +x ./Cardinal

# 运行
./Cardinal
```

### Docker 部署

首先请从 [Docker 官网](https://docs.docker.com) 安装 `docker` 与 `docker-compose`

确保当前用户拥有 `docker` 及 `docker-compose` 权限，然后执行

```bash
curl https://sh.cardinal.ink | bash
```

初次使用应当在下载后配置 `docker-compose.yml` 内的各项参数

## 开始使用

默认端口： `19999`

* 选手端 `http://localhost:19999/`
* 后台管理 `http://localhost:19999/manager`

## 使用 Cardinal 的团队及组织

<a><img src="https://cardinal.ink/brand/QLNU.jpg" height="65px"  style="margin-right: 10px; padding: 10px; border: 1px solid #ccc; border-radius: 5px;"/></a>
<a href="https://cnc.poliupg.ac.id/" target="_blank"><img src="https://cardinal.ink/brand/CNC.png" height="65px" style="margin-right: 10px; padding: 10px; border: 1px solid #ccc; border-radius: 5px;"/></a>

## 开源协议

© Vidar-Team

GNU Affero General Public License v3.0
