[![Cardinal Logo](https://img.cdn.n3ko.co/lsky/2020/02/16/e75b82afd0932.png)](https://cardinal.ink)
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
<!-- ALL-CONTRIBUTORS-BADGE:END -->
[![Build](https://travis-ci.com/vidar-team/Cardinal.svg?branch=master)](https://travis-ci.org/vidar-team/Cardinal)
[![GoReport](https://goreportcard.com/badge/github.com/vidar-team/Cardinal)](https://goreportcard.com/report/github.com/vidar-team/Cardinal)
[![QQ Group](https://img.shields.io/badge/QQ%E7%BE%A4-130818749-blue.svg?logo=Tencent%20QQ)](https://shang.qq.com/wpa/qunwpa?idkey=c6a35c5fbec05fdcd2d2605e08b4b5f8d6e5854471fefd8c03d370d14870b818)
[![Discord](https://img.shields.io/discord/721936261778243615?label=Discord&logo=Discord)](https://discord.gg/F2EfgbM)

# [Cardinal](https://cardinal.ink) —— CTF AWD 线下赛平台
## 介绍

Cardinal 是由 Vidar-Team 开发的 AWD 比赛平台，使用 Go 编写。本程序可以作为 CTF 线下比赛平台，亦可用于团队内部 AWD 模拟练习。

![Cardinal Frontend](https://s1.ax1x.com/2020/05/28/tVPltI.png)

![Cardinal Backend](https://s1.ax1x.com/2020/05/28/tVP1ht.png)

![Asteroid](https://s1.ax1x.com/2020/05/28/tVP6jU.png)
（该 AWD 实时 3D 攻击大屏为项目 [Asteroid](https://github.com/wuhan005/Asteroid)，已适配接入 Cardinal）

## 使用文档
[cardinal.ink](https://cardinal.ink)
> 请您在使用前认真阅读使用文档，谢谢 ♪(･ω･)ﾉ

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

## Contributors ✨

[提交 Bug](https://github.com/vidar-team/Cardinal/issues/new) | [Fork & Pull Request](https://github.com/vidar-team/Cardinal/fork)
<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://github.com/wuhan005"><img src="https://avatars3.githubusercontent.com/u/12731778?v=4" width="100px;" alt=""/><br /><sub><b>E99p1ant</b></sub></a><br /><a href="https://github.com/vidar-team/Cardinal/commits?author=wuhan005" title="Code">💻</a> <a href="#design-wuhan005" title="Design">🎨</a> <a href="https://github.com/vidar-team/Cardinal/commits?author=wuhan005" title="Documentation">📖</a> <a href="#maintenance-wuhan005" title="Maintenance">🚧</a></td>
    <td align="center"><a href="https://github.com/Moesang"><img src="https://avatars2.githubusercontent.com/u/46858006?v=4" width="100px;" alt=""/><br /><sub><b>Moesang</b></sub></a><br /><a href="https://github.com/vidar-team/Cardinal/commits?author=Moesang" title="Code">💻</a> <a href="https://github.com/vidar-team/Cardinal/commits?author=Moesang" title="Documentation">📖</a> <a href="#maintenance-Moesang" title="Maintenance">🚧</a></td>
    <td align="center"><a href="https://github.com/michaelfyc"><img src="https://avatars2.githubusercontent.com/u/45136049?v=4" width="100px;" alt=""/><br /><sub><b>michaelfyc</b></sub></a><br /><a href="#translation-michaelfyc" title="Translation">🌍</a></td>
  </tr>
</table>

<!-- markdownlint-enable -->
<!-- prettier-ignore-end -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

十分欢迎您和我们一起改进 Cardinal，您可以改进现有程序，加入新功能，完善文档，优化代码等。

## 协议与许可

© Vidar-Team

本项目使用 APACHE LICENSE VERSION 2.0 进行许可。

若您安装及/或使用 Cardinal 及其相关软件、文档，即表示您已充分阅读、理解并同意接受本协议。

我们接受并允许各大高校、安全团队、技术爱好者使用 Cardinal 作为比赛训练平台或举办内部训练赛。

但不允许在未经许可授权的情况下，使用 Cardinal 的代码、文档、相关软件等开展商业培训、商业比赛、产品销售等任何营利性行为。禁止恶意更换、去除 Cardinal 及其相关软件、文档版权信息。

一经发现，严肃处理。Vidar-Team 保留追究其法律责任的权力。
