// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package clock

import (
	"math"
	"time"

	"github.com/pkg/errors"

	"github.com/vidar-team/Cardinal/internal/conf"
)

type Status int

const (
	StatusWait Status = iota
	StatusRunning
	StatusPause
	StatusEnd
)

var T = new(Clock)

type Clock struct {
	StartAt             time.Time
	EndAt               time.Time
	RoundDuration       time.Duration
	RestTime            [][]time.Time
	RunTime             [][]time.Time
	TotalRound          uint
	CurrentRound        uint
	RoundRemainDuration time.Duration
	Status              Status

	stopChan chan struct{}
}

func Init() error {
	if conf.Game.RoundDuration == 0 {
		return ErrZeroRoundDuration
	}

	T = &Clock{
		StartAt:       conf.Game.StartAt.In(time.Local),
		EndAt:         conf.Game.EndAt.In(time.Local),
		RoundDuration: time.Duration(conf.Game.RoundDuration) * time.Minute,
	}

	restTime := make([][]time.Time, 0, len(conf.Game.PauseTime))
	for _, t := range conf.Game.PauseTime {
		restTime = append(restTime, []time.Time{t.StartAt.In(time.Local), t.EndAt.In(time.Local)})
	}
	T.RestTime = restTime

	// Check timer configuration.
	if err := T.checkConfig(); err != nil {
		return errors.Wrap(err, "check config")
	}

	T.RestTime = combineDuration(T.RestTime)

	// Set competition run time cycle.
	if len(T.RestTime) != 0 {
		// StartAt -> RestTime[0][Start]
		T.RunTime = append(T.RunTime, []time.Time{T.StartAt, T.RestTime[0][0]})
		for i := 0; i < len(T.RestTime)-1; i++ {
			// Runtime = RestHeadEnd -> RestNextBegin
			T.RunTime = append(T.RunTime, []time.Time{T.RestTime[i][1], T.RestTime[i+1][0]})
		}
		// RestTime[Last][End] -> EndAt
		T.RunTime = append(T.RunTime, []time.Time{T.RestTime[len(T.RestTime)-1][1], T.EndAt})

	} else {
		T.RunTime = append(T.RunTime, []time.Time{T.StartAt, T.EndAt})
	}

	// Calculate total round count.
	var totalTime time.Duration
	for _, duration := range T.RunTime {
		totalTime += duration[1].Sub(duration[0])
	}
	T.TotalRound = uint(math.Ceil(totalTime.Minutes() / T.RoundDuration.Minutes()))

	return nil
}

// checkConfig checks the time configuration from the configuration file.
// It checks the order of StartAt and EndAt, each RestTime.
func (c *Clock) checkConfig() error {
	if c.StartAt.After(c.EndAt) {
		return ErrStartTimeOrder
	}

	// Check rest time.
	for key, duration := range c.RestTime {
		if len(duration) != 2 {
			return ErrRestTimeFormat
		}

		start := duration[0]
		end := duration[1]

		if start.After(end) {
			return ErrRestTimeOrder
		}

		if start.Before(c.StartAt) || end.After(c.EndAt) {
			return ErrRestTimeOverflow
		}

		// RestTime should in order.
		if key != 0 {
			previousStart := c.RestTime[key-1][0]
			if start.Before(previousStart) {
				return ErrRestTimeListOrder
			}
		}
	}

	return nil
}

// combineDuration combines time duration, the operation is idempotent.
// If two time duration are overlapped, the former one will be combined with the latter one,
// and the former duration will be set to nil and be removed.
func combineDuration(d [][]time.Time) [][]time.Time {
	for i := 0; i < len(d)-1; i++ {
		headIndex := i
		nextIndex := i + 1

		headBegin, headEnd := d[headIndex][0], d[headIndex][1]
		nextBegin, nextEnd := d[nextIndex][0], d[nextIndex][1]

		// Head: ... ============== E
		// Next:        B ===...
		if headEnd.After(nextBegin) {
			if headEnd.After(nextEnd) {
				// Head: ... ============ E
				// Next:       B === E
				// 			v
				// Head: ... ============ E
				// Next: ... ============ E
				d[nextIndex] = d[headIndex]
			} else {
				// Head: ... ============ E
				// Next:       B ============= E
				//            v
				// Head: ... ============ E
				// Next: ... ================= E
				d[nextIndex][0] = headBegin
			}
			d[headIndex] = nil
		}
	}

	// Remove the empty rest time element.
	for i := 0; i < len(d); i++ {
		if d[i] == nil {
			d = append(d[:i], d[i+1:]...)
		}
	}

	return d
}
