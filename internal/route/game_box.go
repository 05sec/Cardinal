// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/form"
	"github.com/vidar-team/Cardinal/internal/i18n"
)

// GameBoxHandler is the game box request handler.
type GameBoxHandler struct{}

// NewGameBoxHandler creates and returns a new game box handler.
func NewGameBoxHandler() *GameBoxHandler {
	return &GameBoxHandler{}
}

// List returns all the game boxes.
func (*GameBoxHandler) List(ctx context.Context) error {
	gameBoxes, err := db.GameBoxes.Get(ctx.Request().Context(), db.GetGameBoxesOption{})
	if err != nil {
		log.Error("Failed to get game box list: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success(gameBoxes)
}

// New creates game boxes with the given options.
func (*GameBoxHandler) New(ctx context.Context, f form.NewGameBox, l *i18n.Locale) error {
	gameBoxOptions := make([]db.CreateGameBoxOptions, 0, len(f))
	for _, option := range f {
		gameBoxOptions = append(gameBoxOptions, db.CreateGameBoxOptions{
			TeamID:      option.TeamID,
			ChallengeID: option.ChallengeID,
			Address:     option.Address,
			Description: option.Description,
			InternalSSH: db.SSHConfig{
				Port:     option.SSHPort,
				User:     option.SSHUser,
				Password: option.SSHPassword,
			},
		})
	}

	_, err := db.GameBoxes.BatchCreate(ctx.Request().Context(), gameBoxOptions)
	if err != nil {
		if err == db.ErrGameBoxAlreadyExists {
			// TODO show which game box has existed.
			return ctx.Error(40000, l.T("gamebox.repeat"))
		}
		log.Error("Failed to create game boxes in batch: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success(gameBoxOptions)
}

// Update updates the game box.
func (*GameBoxHandler) Update(ctx context.Context, f form.UpdateGameBox, l *i18n.Locale) error {
	err := db.GameBoxes.Update(ctx.Request().Context(), f.ID, db.UpdateGameBoxOptions{
		Address:     f.Address,
		Description: f.Description,
		InternalSSH: db.SSHConfig{
			Port:     f.SSHPort,
			User:     f.SSHUser,
			Password: f.SSHPassword,
		},
	})
	if err == db.ErrGameBoxNotExists {
		return ctx.Error(40400, "gamebox.not_found")
	}
	return ctx.Success()
}

// ResetAll resets all the game boxes.
// It deletes all the game boxes score record and refresh the ranking list.
func (*GameBoxHandler) ResetAll(ctx context.Context) error {
	// TODO
	return nil
}

// SSHTest tests the game box SSH configuration,
// which try to connect to the game box instance within SSH.
func (*GameBoxHandler) SSHTest(ctx context.Context) error {
	// TODO
	return nil
}

// RefreshFlag refreshes the game box flag if the `RenewFlagCommand` was set in challenge.
// It will connect to the game box instance and run the command to refresh the flag.
func (*GameBoxHandler) RefreshFlag(ctx context.Context) error {
	// TODO
	return nil
}
