// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"

	"github.com/thanhpk/randstr"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/form"
	"github.com/vidar-team/Cardinal/internal/i18n"
)

type TeamHandler struct{}

func NewTeamHandler() *TeamHandler {
	return &TeamHandler{}
}

// List returns all the teams.
func (*TeamHandler) List(ctx context.Context) error {
	teams, err := db.Teams.Get(ctx.Request().Context(), db.GetTeamsOptions{})
	if err != nil {
		log.Error("Failed to get teams: %v", err)
		return ctx.ServerError()
	}

	type team struct {
		ID    uint    `json:"ID"`
		Logo  string  `json:"Logo"`
		Score float64 `json:"Score"`
		Rank  uint    `json:"Rank"`
		Token string  `json:"Token"`
	}

	teamList := make([]*team, 0, len(teams))

	for _, t := range teams {
		teamList = append(teamList, &team{
			ID:    t.ID,
			Logo:  t.Logo,
			Score: t.Score,
			Rank:  t.Rank,
			Token: t.Token,
		})
	}

	return ctx.Success(teamList)
}

// New creates a new team with the given options.
func (*TeamHandler) New(ctx context.Context, f form.NewTeam, l *i18n.Locale) error {
	teamOptions := make([]db.CreateTeamOptions, 0, len(f))
	for _, team := range f {
		password := randstr.String(16)
		teamOptions = append(teamOptions, db.CreateTeamOptions{
			Name:     team.Name,
			Password: password,
			Logo:     team.Logo,
		})
	}

	_, err := db.Teams.BatchCreate(ctx.Request().Context(), teamOptions)
	if err != nil {
		if err == db.ErrTeamAlreadyExists {
			// TODO show which team has existed.
			return ctx.Error(40000, l.T("team.repeat"))
		}
		log.Error("Failed to create teams in batch: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success(teamOptions)
}

// Update updates the team with the given options.
func (*TeamHandler) Update(ctx context.Context, f form.UpdateTeam, l *i18n.Locale) error {
	// Check the team exist or not.
	team, err := db.Teams.GetByID(ctx.Request().Context(), f.ID)
	if err != nil {
		if err == db.ErrTeamNotExists {
			return ctx.Error(40400, l.T("team.not_found"))
		}
		log.Error("Failed to get team by ID: %v", err)
		return ctx.ServerError()
	}

	newTeam, err := db.Teams.GetByName(ctx.Request().Context(), f.Name)
	if err == nil {
		if team.ID != newTeam.ID {
			// TODO i18n
			return ctx.Error(40000, fmt.Sprintf("Team name %q repeat.", team.Name))
		}
	} else if err != db.ErrTeamNotExists {
		log.Error("Failed to get team by name: %v", err)
		return ctx.ServerError()
	}

	err = db.Teams.Update(ctx.Request().Context(), f.ID, db.UpdateTeamOptions{
		Name: f.Name,
		Logo: f.Logo,
	})
	if err != nil {
		log.Error("Failed to update team: %v", err)
		return ctx.ServerError()
	}
	return ctx.Success()
}

// Delete deletes the team with the given ID.
func (*TeamHandler) Delete(ctx context.Context, l *i18n.Locale) error {
	id := uint(ctx.QueryInt("id"))

	// Check the team exist or not.
	team, err := db.Teams.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err == db.ErrTeamNotExists {
			return ctx.Error(40400, l.T("team.not_found"))
		}
		log.Error("Failed to get team by ID: %v", err)
		return ctx.ServerError()
	}

	err = db.Teams.DeleteByID(ctx.Request().Context(), team.ID)
	if err != nil {
		log.Error("Failed to delete team: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success()
}

// ResetPassword resets team password with the given id.
func (*TeamHandler) ResetPassword(ctx context.Context, l *i18n.Locale) error {
	id := uint(ctx.QueryInt("id"))

	// Check the team exist or not.
	team, err := db.Teams.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err == db.ErrTeamNotExists {
			return ctx.Error(40400, l.T("team.not_found"))
		}
		log.Error("Failed to get team by ID: %v", err)
		return ctx.ServerError()
	}

	newPassword := randstr.String(16)
	err = db.Teams.ChangePassword(ctx.Request().Context(), team.ID, newPassword)
	if err != nil {
		log.Error("Failed to change password: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success(newPassword)
}
