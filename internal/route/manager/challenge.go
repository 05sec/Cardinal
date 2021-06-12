// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package manager

import (
	"time"

	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/form"
)

// GetChallenges returns all the challenges.
func (*Handler) GetChallenges(ctx context.Context) error {
	type challenge struct {
		ID               uint      `json:"ID"`
		CreatedAt        time.Time `json:"CreatedAt"`
		Title            string    `json:"Title"`
		Visible          bool      `json:"Visible"`
		BaseScore        float64   `json:"BaseScore"`
		AutoRenewFlag    bool      `json:"AutoRenewFlag"`
		RenewFlagCommand string    `json:"RenewFlagCommand"`
	}

	challenges, err := db.Challenges.Get(ctx.Request().Context())
	if err != nil {
		log.Error("Failed to get challenges: %v", err)
		return ctx.ServerError()
	}

	challengeList := make([]*challenge, 0, len(challenges))
	for _, c := range challenges {
		// The challenge visible value depends on the challenge's game boxes' visible.
		// So we get one of the game box of the challenge and use its visible data.
		gameBoxes, err := db.GameBoxes.Get(ctx.Request().Context(), db.GetGameBoxesOption{
			ChallengeID: c.ID,
		})
		if err != nil {
			log.Error("Failed to get game boxes list: %v", err)
			return ctx.ServerError()
		}
		var challengeVisible bool
		if len(gameBoxes) != 0 {
			challengeVisible = gameBoxes[0].Visible
		}

		challengeList = append(challengeList, &challenge{
			ID:               c.ID,
			CreatedAt:        c.CreatedAt,
			Title:            c.Title,
			Visible:          challengeVisible,
			BaseScore:        c.BaseScore,
			AutoRenewFlag:    c.AutoRenewFlag,
			RenewFlagCommand: c.RenewFlagCommand,
		})
	}

	return ctx.Success(challengeList)
}

// NewChallenge creates a new challenge.
func (*Handler) NewChallenge(ctx context.Context, f form.NewChallenge) error {
	_, err := db.Challenges.Create(ctx.Request().Context(), db.CreateChallengeOptions{
		Title:            f.Title,
		BaseScore:        f.BaseScore,
		AutoRenewFlag:    f.AutoRenewFlag,
		RenewFlagCommand: f.RenewFlagCommand,
	})
	if err != nil {
		log.Error("Failed to create new challenge: %v", err)
		return ctx.ServerError()
	}

	// TODO send log to panel

	// TODO i18n
	return ctx.Success("Success")
}

// UpdateChallenge updates the challenge with the given ID.
func (*Handler) UpdateChallenge(ctx context.Context, f form.UpdateChallenge) error {
	// Check if the challenge exists.
	_, err := db.Challenges.GetByID(ctx.Request().Context(), f.ID)
	if err != nil {
		if err == db.ErrChallengeNotExists {
			// TODO i18n
			return ctx.Error(40000, "Challenge does not exist.")
		}
		log.Error("Failed to get challenge: %v", err)
		return ctx.ServerError()
	}

	err = db.Challenges.Update(ctx.Request().Context(), f.ID, db.UpdateChallengeOptions{
		Title:            f.Title,
		BaseScore:        f.BaseScore,
		AutoRenewFlag:    f.AutoRenewFlag,
		RenewFlagCommand: f.RenewFlagCommand,
	})
	if err != nil {
		log.Error("Failed to update challenge: %v", err)
		return ctx.ServerError()
	}
	// TODO i18n
	return ctx.Success("Success")
}
