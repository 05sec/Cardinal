package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math"
	"strconv"
	"time"
)

// Timer is the time data of the Cardinal.
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
	s.checkTimeConfig()

	// Calculate the rest time cycle.
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
	// Remove the empty element.
	for i := 0; i < len(s.Timer.RestTime); i++ {
		if s.Timer.RestTime[i] == nil {
			s.Timer.RestTime = append(s.Timer.RestTime[:i], s.Timer.RestTime[i+1:]...)
			i++
		}
	}

	// Set the competition time cycle.
	if len(s.Timer.RestTime) != 0 {
		s.Timer.RunTime = append(s.Timer.RunTime, []time.Time{s.Timer.BeginTime, s.Timer.RestTime[0][0]})
		for i := 0; i < len(s.Timer.RestTime)-1; i++ {
			s.Timer.RunTime = append(s.Timer.RunTime, []time.Time{s.Timer.RestTime[i][1], s.Timer.RestTime[i+1][0]})
		}
		s.Timer.RunTime = append(s.Timer.RunTime, []time.Time{s.Timer.RestTime[len(s.Timer.RestTime)-1][1], s.Timer.EndTime})

	} else {
		s.Timer.RunTime = append(s.Timer.RunTime, []time.Time{s.Timer.BeginTime, s.Timer.EndTime})
	}

	// Calculate the total round.
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
	beginTime := s.Timer.BeginTime.Unix()
	endTime := s.Timer.EndTime.Unix()
	lastRoundCalculate := false // A sign for the last round score calculate.

	{
		s.SetRankListTitle() // Refresh ranking list table header.
		s.SetRankList()
	}

	for {
		nowTime := time.Now().Unix()
		s.Timer.RoundRemainTime = -1

		if nowTime > beginTime && nowTime < endTime {
			nowRunTimeIndex := -1
			for index, dur := range s.Timer.RunTime {
				if nowTime > dur[0].Unix() && nowTime < dur[1].Unix() {
					nowRunTimeIndex = index // Get which time cycle now.
					break
				}
			}

			if nowRunTimeIndex == -1 {
				// Suspended
				s.Timer.Status = "pause"
			} else {
				// In progress
				s.Timer.Status = "on"
				var nowRound int
				var workTime int64 // Cumulative time until now.

				for index, dur := range s.Timer.RunTime {
					if index < nowRunTimeIndex {
						workTime += dur[1].Unix() - dur[0].Unix()
					} else {
						workTime += nowTime - dur[0].Unix()
						break
					}
				}
				nowRound = int(math.Ceil(float64(workTime) / float64(s.Timer.Duration*60))) // Calculate current round.
				s.Timer.RoundRemainTime = nowRound*int(s.Timer.Duration)*60 - int(workTime) // Calculate the time to next round.

				// Check if it is a new round.
				if s.Timer.NowRound < nowRound {
					s.Timer.NowRound = nowRound
					// New round hook

					// Clean the status of the gameboxes.
					s.Mysql.Model(&GameBox{}).Update(map[string]interface{}{"is_down": false, "is_attacked": false})

					// Calculate scores.
					// Get the latest score record.
					var latestScore Score
					s.Mysql.Model(&Score{}).Order("`round` DESC").Limit(1).Find(&latestScore)

					// If Cardinal has been restart by unexpected error, get the latest round score and chick if need calculate the scores of previous round.
					if latestScore.Round < s.Timer.NowRound-1 {
						go s.CalculateRoundScore(s.Timer.NowRound - 1)
					}

					fmt.Println(s.Timer.NowRound)
				}
			}

		} else if nowTime < beginTime {
			// Not started.
			s.Timer.Status = "wait"
		} else {
			// Over.
			// Calculate the score of the last round when the competition is over.
			if !lastRoundCalculate {
				lastRoundCalculate = true
				go s.CalculateRoundScore(s.Timer.TotalRound)
				s.NewLog(IMPORTANT, "system", "比赛已结束")
			}

			s.Timer.Status = "end"
		}

		time.Sleep(1 * time.Second)
	}
}

func (s *Service) checkTimeConfig() {
	if s.Timer.BeginTime.Unix() > s.Timer.EndTime.Unix() {
		log.Fatalln("比赛结束时间应大于开始时间！")
	}

	// Check the RestTime in config file is correct.
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
