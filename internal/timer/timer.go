package timer

import (
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/asteroid"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/misc/webhook"
	"github.com/vidar-team/Cardinal/internal/utils"
	log "unknwon.dev/clog/v2"
)

var t = new(timer)

// timer is the time data struct of the Cardinal.
type timer struct {
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

// Get returns the timer.
func Get() *timer {
	return t
}

// GetTime is the HTTP Handler of the time.
func GetTime(c *gin.Context) (int, interface{}) {
	return utils.MakeSuccessJSON(gin.H{
		"BeginTime":       t.BeginTime.Unix(),
		"EndTime":         t.EndTime.Unix(),
		"Duration":        t.Duration,
		"NowRound":        t.NowRound,
		"NowTime":         time.Now().Unix(),
		"RoundRemainTime": t.RoundRemainTime,
		"Status":          t.Status,
	})
}

func Init() {
	// Check the bridge.
	if SetRankListTitle == nil ||
		SetRankList == nil ||
		CleanGameBoxStatus == nil ||
		GetLatestScoreRound == nil ||
		RefreshFlag == nil ||
		CalculateRoundScore == nil {

		log.Fatal("Timer bridge error, the function should be not nil.")
	}

	t = &timer{
		BeginTime: conf.Get().BeginTime,
		EndTime:   conf.Get().EndTime,
		Duration:  conf.Get().Duration,
		RestTime:  conf.Get().RestTime,
		NowRound:  -1,
	}
	checkTimeConfig()

	// Calculate the rest time cycle.
	for i := 0; i < len(t.RestTime)-1; i++ {
		j := i + 1
		if t.RestTime[i][1].Unix() >= t.RestTime[j][0].Unix() {
			if t.RestTime[i][1].Unix() >= t.RestTime[j][1].Unix() {
				t.RestTime[j] = t.RestTime[i]
			} else {
				t.RestTime[j][0] = t.RestTime[i][0]
			}
			t.RestTime[i] = nil
		} else {
			i++
		}
	}

	// Remove the empty element.
	for i := 0; i < len(t.RestTime); i++ {
		if t.RestTime[i] == nil {
			t.RestTime = append(t.RestTime[:i], t.RestTime[i+1:]...)
			i++
		}
	}

	// Set the competition time cycle.
	if len(t.RestTime) != 0 {
		t.RunTime = append(t.RunTime, []time.Time{t.BeginTime, t.RestTime[0][0]})
		for i := 0; i < len(t.RestTime)-1; i++ {
			t.RunTime = append(t.RunTime, []time.Time{t.RestTime[i][1], t.RestTime[i+1][0]})
		}
		t.RunTime = append(t.RunTime, []time.Time{t.RestTime[len(t.RestTime)-1][1], t.EndTime})

	} else {
		t.RunTime = append(t.RunTime, []time.Time{t.BeginTime, t.EndTime})
	}

	// Calculate the total round.
	var totalTime int64
	for _, dur := range t.RunTime {
		totalTime += dur[1].Unix() - dur[0].Unix()
	}
	t.TotalRound = int(totalTime / 60 / int64(t.Duration))

	log.Trace(string(locales.I18n.T(conf.Get().SystemLanguage, "timer.total_round", gin.H{
		"round": t.TotalRound,
	})))

	log.Trace(string(locales.I18n.T(conf.Get().SystemLanguage, "timer.total_time", gin.H{
		"time": int(totalTime / 60),
	})))

	go timerProcess()
}

func timerProcess() {
	beginTime := t.BeginTime.Unix()
	endTime := t.EndTime.Unix()
	lastRoundCalculate := false // A sign for the last round score calculate.

	{
		SetRankListTitle() // Refresh ranking list table header.
		SetRankList()
	}

	for {
		nowTime := time.Now().Unix()
		t.RoundRemainTime = -1

		if nowTime > beginTime && nowTime < endTime {
			nowRunTimeIndex := -1
			for index, dur := range t.RunTime {
				if nowTime > dur[0].Unix() && nowTime < dur[1].Unix() {
					nowRunTimeIndex = index // Get which time cycle now.
					break
				}
			}

			if nowRunTimeIndex == -1 {
				// Suspended
				if t.Status != "pause" {
					go webhook.Add(webhook.PAUSE_HOOK, nil)
				}
				t.Status = "pause"
			} else {
				// In progress
				t.Status = "on"
				var nowRound int
				var workTime int64 // Cumulative time until now.

				for index, dur := range t.RunTime {
					if index < nowRunTimeIndex {
						workTime += dur[1].Unix() - dur[0].Unix()
					} else {
						workTime += nowTime - dur[0].Unix()
						break
					}
				}
				nowRound = int(math.Ceil(float64(workTime) / float64(t.Duration*60))) // Calculate current round.
				t.RoundRemainTime = nowRound*int(t.Duration)*60 - int(workTime)       // Calculate the time to next round.

				// Check if it is a new round.
				if t.NowRound < nowRound {
					t.NowRound = nowRound
					if t.NowRound == 1 {
						// Game start hook
						go webhook.Add(webhook.BEGIN_HOOK, nil)
					}

					// New round hook
					go webhook.Add(webhook.BEGIN_HOOK, t.NowRound)

					// Clean the status of the gameboxes.
					CleanGameBoxStatus()
					SetRankList()

					// Calculate scores.
					// Get the latest score record.
					latestScoreRound := GetLatestScoreRound()

					// If Cardinal has been restart by unexpected error, get the latest round score and chick if need calculate the scores of previous round.
					if latestScoreRound < t.NowRound-1 {
						CalculateRoundScore(t.NowRound - 1)
					}

					// Auto refresh flag
					RefreshFlag()

					// Asteroid Unity3D refresh.
					asteroid.NewRoundAction()

					log.Trace("New round: %d", t.NowRound)
				}
			}

		} else if nowTime < beginTime {
			// Not started.
			t.Status = "wait"
		} else {
			// Over.
			// Calculate the score of the last round when the competition is over.
			if !lastRoundCalculate {
				lastRoundCalculate = true
				go CalculateRoundScore(t.TotalRound)
				// Game over hook
				go webhook.Add(webhook.END_HOOK, nil)
				logger.New(logger.IMPORTANT, "system", string(locales.I18n.T(conf.Get().SystemLanguage, "timer.end")))
			}

			t.Status = "end"
		}

		time.Sleep(1 * time.Second)
	}
}

func checkTimeConfig() {
	if t.BeginTime.Unix() > t.EndTime.Unix() {
		log.Fatal(string(locales.I18n.T(conf.Get().SystemLanguage, "timer.start_time_error")))
	}

	// Check the RestTime in config file is correct.
	for key, dur := range t.RestTime {
		if len(dur) != 2 {
			log.Fatal(string(locales.I18n.T(conf.Get().SystemLanguage, "timer.single_rest_time_error")))
		}
		if dur[0].Unix() >= dur[1].Unix() {
			log.Fatal(string(locales.I18n.T(conf.Get().SystemLanguage, "timer.rest_time_start_error",
				gin.H{
					"from": dur[0].String(),
					"to":   dur[1].String(),
				},
			)))
		}
		if dur[0].Unix() <= t.BeginTime.Unix() || dur[1].Unix() >= t.EndTime.Unix() {
			log.Fatal(string(locales.I18n.T(conf.Get().SystemLanguage, "timer.rest_time_overflow_error",
				gin.H{
					"from": dur[0].String(),
					"to":   dur[1].String(),
				},
			)))
		}
		// RestTime should in order.
		if key != 0 && dur[0].Unix() <= t.RestTime[key-1][0].Unix() {
			log.Fatal(string(locales.I18n.T(conf.Get().SystemLanguage, "timer.rest_time_order_error",
				gin.H{
					"from": dur[0].String(),
					"to":   dur[1].String(),
				},
			)))
		}
	}
}
