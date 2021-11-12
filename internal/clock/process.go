// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package clock

import (
	"context"
	"math"
	"sync"
	"time"

	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/misc/webhook"
	"github.com/vidar-team/Cardinal/internal/rank"
	//"github.com/vidar-team/Cardinal/internal/score"
)

// Start starts the game clock processor routine.
func Start() {
	// TODO: only one timer started.
	go T.start()
}

// Stop stops the game clock processor.
func Stop() {
	T.stopChan <- struct{}{}
}

func (c *Clock) start() {
	ctx, cancel := context.WithCancel(context.Background())

	// Refresh the ranking list.
	if err := rank.SetTitle(ctx); err != nil {
		log.Error("Failed to set rank title: %v", err)
	}
	if err := rank.SetRankList(ctx); err != nil {
		log.Error("Failed to set rank list: %v", err)
	}

	var latestCalculateRound uint
	lastRound := sync.Once{}

	for {
		select {
		case <-c.stopChan:
			cancel()
			close(c.stopChan)
			return

		default:
			currentTime := time.Now()

			if currentTime.Before(c.StartAt) {
				// The game is not started.
				c.Status = StatusWait
				continue

			} else if currentTime.After(c.EndAt) {
				// The game is over.
				// Calculate the score of the last round when the game is over.
				lastRound.Do(func() {
					if err := db.Scores.Calculate(ctx, c.TotalRound); err != nil {
						log.Error("Failed to calculate the last round score: %v", err)
					}

					go webhook.Add(webhook.END_HOOK, nil)
					// TODO logger.New(logger.IMPORTANT, "system", locales.T("timer.end"))
				})

				c.Status = StatusEnd
				continue
			}

			// The game is running.
			// Get which time cycle now.
			currentRunTimeIndex := -1
			for index, duration := range c.RunTime {
				if currentTime.After(duration[0]) && currentTime.Before(duration[1]) {
					currentRunTimeIndex = index
					break
				}
			}

			if currentRunTimeIndex == -1 {
				// Suspended
				if c.Status != StatusPause {
					go webhook.Add(webhook.PAUSE_HOOK, nil)
				}
				c.Status = StatusPause

			} else {
				// In progress
				c.Status = StatusRunning

				// Cumulative time until now.
				var runningDuration time.Duration
				for index, duration := range c.RunTime {
					if index < currentRunTimeIndex {
						runningDuration += duration[1].Sub(duration[0])
					} else {
						// The last runtime cycle for now.
						runningDuration += currentTime.Sub(duration[0])
						break
					}
				}

				// Get current round.
				currentRound := uint(math.Ceil(runningDuration.Seconds() / c.RoundDuration.Seconds()))
				// Calculate the time duration next round.
				c.RoundRemainDuration = time.Duration(currentRound)*c.RoundDuration - runningDuration

				// Check if it is a new round.
				if c.CurrentRound < currentRound {
					c.CurrentRound = currentRound
					if c.CurrentRound == 1 {
						go webhook.Add(webhook.BEGIN_HOOK, nil)
					}

					go webhook.Add(webhook.BEGIN_HOOK, c.CurrentRound)

					// Clean the status of the game boxes.
					if err := db.GameBoxes.CleanAllStatus(ctx); err != nil {
						log.Error("Failed to clean game boxes' status: %v", err)
					}

					// Refresh the ranking list.
					if err := rank.SetTitle(ctx); err != nil {
						log.Error("Failed to set rank title: %v", err)
					}
					if err := rank.SetRankList(ctx); err != nil {
						log.Error("Failed to set rank list: %v", err)
					}

					// If Cardinal has been restart by accident, get the latest round score and chick if it needs to calculate the scores of previous round.
					// The default value of `latestCalculateRound` is 0, it means that Cardinal will calculate the score when started.
					if latestCalculateRound < c.CurrentRound-1 {
						if err := db.Scores.Calculate(ctx, c.CurrentRound-1); err != nil {
							log.Error("Failed to calculate score: %v", err)
						}
						latestCalculateRound = c.CurrentRound - 1
					}

					// TODO Auto refresh flag
					//RefreshFlag()

					// TODO Asteroid Unity3D refresh.
					//asteroid.NewRoundAction()
				}
			}

			time.Sleep(1 * time.Second)
		}
	}
}
