package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"strconv"
	"time"
)

type Timer struct {
	BeginTime       time.Time     // init
	EndTime         time.Time     // init
	Duration        uint          // init
	RestTime        [][]time.Time // init
	RunTime         [][]time.Time // init
	TotalRound      int           // init
	NowRound        int
	RoundRemainTime int
	Status          string
}

func (s *Service) getTime() (int, interface{}) {
	return s.makeSuccessJSON(gin.H{
		"BeginTime":       s.Timer.BeginTime.Unix(),
		"EndTime":         s.Timer.EndTime.Unix(),
		"Duration":        s.Timer.Duration,
		"NowRound":        s.Timer.NowRound,
		"RoundRemainTime": s.Timer.RoundRemainTime,
		"Status":          s.Timer.Status,
	})
}

func (s *Service) initTimer() {
	s.Timer = &Timer{
		BeginTime: s.Conf.Base.BeginTime,
		EndTime:   s.Conf.Base.EndTime,
		Duration:  s.Conf.Base.Duration,
		RestTime:  s.Conf.Base.RestTime,
		NowRound:  -1,
	}
	s.checkRestTimeConfig()

	// 计算休息时间周期
	for i := 0; i < len(s.Timer.RestTime)-1; i++ {
		j := i + 1
		if s.Timer.RestTime[i][1].Unix() >= s.Timer.RestTime[j][0].Unix() {
			if s.Timer.RestTime[i][1].Unix() >= s.Timer.RestTime[j][1].Unix() {
				s.Timer.RestTime[j] = s.Timer.RestTime[i]
			} else {
				s.Timer.RestTime[j][0] = s.Timer.RestTime[i][0]
			}
			s.Timer.RestTime[i] = nil
		} else {
			i++
		}
	}
	// 去除空元素
	for i := 0; i < len(s.Timer.RestTime); i++ {
		if s.Timer.RestTime[i] == nil {
			s.Timer.RestTime = append(s.Timer.RestTime[:i], s.Timer.RestTime[i+1:]...)
			i++
		}
	}

	// 设置 RunTime 比赛时间周期
	if len(s.Timer.RestTime) != 0 {
		s.Timer.RunTime = append(s.Timer.RunTime, []time.Time{s.Timer.BeginTime, s.Timer.RestTime[0][0]})
		for i := 0; i < len(s.Timer.RestTime)-1; i++ {
			s.Timer.RunTime = append(s.Timer.RunTime, []time.Time{s.Timer.RestTime[i][1], s.Timer.RestTime[i+1][0]})
		}
		s.Timer.RunTime = append(s.Timer.RunTime, []time.Time{s.Timer.RestTime[len(s.Timer.RestTime)-1][1], s.Timer.EndTime})

	} else {
		s.Timer.RunTime = append(s.Timer.RunTime, []time.Time{s.Timer.BeginTime, s.Timer.EndTime})
	}

	// 计算总轮数
	var totalTime int64
	for _, dur := range s.Timer.RunTime {
		totalTime += dur[1].Unix() - dur[0].Unix()
	}
	s.Timer.TotalRound = int(totalTime / 60 / int64(s.Timer.Duration))

	log.Println("比赛总轮数：" + strconv.Itoa(s.Timer.TotalRound))
	log.Println("比赛总时长：" + strconv.Itoa(int(totalTime/60)) + " 分钟")

	go s.timerProcess()
}

func (s *Service) timerProcess() {
	// 时间处理协程
	beginTime := s.Timer.BeginTime.Unix()
	endTime := s.Timer.EndTime.Unix()
	lastRoundCalculate := false // 最后一轮结束计算分数

	{
		s.SetRankListTitle() // 刷新排行榜标题
	}

	for {
		nowTime := time.Now().Unix()
		s.Timer.RoundRemainTime = -1

		if nowTime > beginTime && nowTime < endTime {
			nowRunTimeIndex := -1
			for index, dur := range s.Timer.RunTime {
				if nowTime > dur[0].Unix() && nowTime < dur[1].Unix() {
					nowRunTimeIndex = index // 顺便记录当前是在哪个时间周期内的
					break
				}
			}

			if nowRunTimeIndex == -1 {
				// 比赛已暂停
				s.Timer.Status = "pause"
			} else {
				// 比赛进行中
				s.Timer.Status = "on"
				var nowRound int
				var workTime int64 // 比赛进行的累计时间

				for index, dur := range s.Timer.RunTime {
					if index < nowRunTimeIndex {
						workTime += dur[1].Unix() - dur[0].Unix()
					} else {
						workTime += nowTime - dur[0].Unix()
						break
					}
				}
				nowRound = int(math.Ceil(float64(workTime) / float64(s.Timer.Duration*60))) // 计算当前轮数
				s.Timer.RoundRemainTime = nowRound*int(s.Timer.Duration)*60 - int(workTime) // 计算距离下一轮开始的秒数

				// 判断是否进入新一轮
				if s.Timer.NowRound < nowRound {
					s.Timer.NowRound = nowRound
					// 新一轮 Hook
					// 清空靶机状态
					s.Mysql.Model(&GameBox{}).Update(map[string]interface{}{"is_down": false, "is_attacked": false})
					// 计算分数
					go s.NewRoundCalculateScore()
					fmt.Println(s.Timer.NowRound)
				}
			}

		} else if nowTime < beginTime {
			// 比赛未开始
			s.Timer.Status = "wait"
		} else {
			// 比赛已结束
			// 最后一轮结束后结算分数
			if !lastRoundCalculate {
				lastRoundCalculate = true
				s.Timer.NowRound = s.Timer.TotalRound + 1		// 设置当前轮为总轮数 +1，使得可以计算最后一轮分数
				go s.NewRoundCalculateScore()
				s.NewLog(IMPORTANT, "system", "比赛已结束")
			}

			s.Timer.Status = "end"
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *Service) checkRestTimeConfig(){
	// 检查 RestTime 数据是否正确
	for key, dur := range s.Timer.RestTime {
		if len(dur) != 2 {
			log.Fatalln("RestTime 单个时间周期配置错误！")
		}
		if dur[0].Unix() >= dur[1].Unix() {
			log.Fatalln("RestTime 配置错误！前一时间应在后一时间点之前。[ " + dur[0].String() + " - " + dur[1].String() + " ]")
		}
		if dur[0].Unix() <= s.Timer.BeginTime.Unix() || dur[1].Unix() >= s.Timer.EndTime.Unix() {
			log.Fatalln("RestTime 配置错误！不能在比赛开始时间之前或比赛结束时间之后。[ " + dur[0].String() + " - " + dur[1].String() + " ]")
		}
		// 配置数据按开始时间顺序输入，方便后面计算
		if key != 0 && dur[0].Unix() <= s.Timer.RestTime[key-1][0].Unix() {
			log.Fatalln("RestTime 需要按开始时间顺序输入！[ " + dur[0].String() + " - " + dur[1].String() + " ]")
		}
	}
}