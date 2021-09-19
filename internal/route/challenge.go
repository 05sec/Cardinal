// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"time"

	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/form"
	"github.com/vidar-team/Cardinal/internal/i18n"
)

type ChallengeHandler struct{}

func NewChallengeHandler() *ChallengeHandler {
	return &ChallengeHandler{}
}

// List returns all the challenges.
func (*ChallengeHandler) List(ctx context.Context) error {
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

// New creates a new challenge.
func (*ChallengeHandler) New(ctx context.Context, f form.NewChallenge, l *i18n.Locale) error {
	_, err := db.Challenges.Create(ctx.Request().Context(), db.CreateChallengeOptions{
		Title:            f.Title,
		BaseScore:        f.BaseScore,
		AutoRenewFlag:    f.AutoRenewFlag,
		RenewFlagCommand: f.RenewFlagCommand,
	})
	if err != nil {
		if err == db.ErrChallengeAlreadyExists {
			return ctx.Error(40000, l.T("challenge.repeat"))
		}
		log.Error("Failed to create new challenge: %v", err)
		return ctx.ServerError()
	}

	// TODO send log to panel

	return ctx.Success(l.T("challenge.success", f.Title))
}

// Update updates the challenge with the given ID.
func (*ChallengeHandler) Update(ctx context.Context, f form.UpdateChallenge, l *i18n.Locale) error {
	// Check if the challenge exists.
	_, err := db.Challenges.GetByID(ctx.Request().Context(), f.ID)
	if err != nil {
		if err == db.ErrChallengeNotExists {
			return ctx.Error(40400, l.T("challenge.not_found"))
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
	return ctx.Success()
}

// Delete deletes the challenge with the given ID.
func (*ChallengeHandler) Delete(ctx context.Context, l *i18n.Locale) error {
	id := uint(ctx.QueryInt("id"))

	// Check if the challenge exists.
	_, err := db.Challenges.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err == db.ErrChallengeNotExists {
			return ctx.Error(40400, l.T("challenge.not_found"))
		}
		log.Error("Failed to get challenge: %v", err)
		return ctx.ServerError()
	}

	err = db.Challenges.DeleteByID(ctx.Request().Context(), id)
	if err != nil {
		log.Error("Failed to delete challenge: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success()
}

// SetVisible sets the challenge's visible.
func (*ChallengeHandler) SetVisible(ctx context.Context, f form.SetChallengeVisible, l *i18n.Locale) error {
	// Check if the challenge exists.
	challenge, err := db.Challenges.GetByID(ctx.Request().Context(), f.ID)
	if err != nil {
		if err == db.ErrChallengeNotExists {
			return ctx.Error(40400, l.T("challenge.not_found"))
		}
		log.Error("Failed to get challenge: %v", err)
		return ctx.ServerError()
	}

	// Get all the game boxes which belong to this challenge.
	gameBoxes, err := db.GameBoxes.Get(ctx.Request().Context(), db.GetGameBoxesOption{
		ChallengeID: challenge.ID,
	})
	if err != nil {
		log.Error("Failed to get game boxes list: %v", err)
		return ctx.ServerError()
	}

	for _, gameBox := range gameBoxes {
		err := db.GameBoxes.SetVisible(ctx.Request().Context(), gameBox.ID, f.Visible)
		if err != nil {
			log.Error("Failed to set game box visible: %v", err)
			return ctx.ServerError()
		}
	}

	// TODO i18n
	return ctx.Success("challenge.set_visible", challenge.Title)
}
