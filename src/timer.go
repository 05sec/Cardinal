package main

import (
	"math"
	"time"
)

type Timer struct {
	BeginTime     time.Time
	EndTime       time.Time
	Duration      uint
	NowRound      int
	NextRoundTime time.Time
}

func (s *Service) initTimer() {
	s.Timer = &Timer{
		BeginTime: s.Conf.Base.BeginTime,
		EndTime:   s.Conf.Base.EndTime,
		Duration:  s.Conf.Base.Duration,
		NowRound:  -1,
	}

	go s.timerProcess()
}

func (s *Service) timerProcess() {
	// 时间处理协程
	beginTime := s.Timer.BeginTime.Unix()
	endTime := s.Timer.EndTime.Unix()
	for {
		var nowRound int

		// 计算当前轮数
		nowTime := time.Now().Unix()

		if nowTime > beginTime && nowTime < endTime{
			nowRound = int(math.Floor(float64(nowTime-beginTime) / float64(s.Timer.Duration*60)))

			// 判断是否进入新一轮
			if s.Timer.NowRound < nowRound {
				s.Timer.NowRound = nowRound
				// 新一轮 Hook

			}
		}

		time.Sleep(1 * time.Second)
	}
}
