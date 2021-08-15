// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
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
		Rank  uint    `json:"rank"`
		Token string  `json:"token"`
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
	type teamInfo struct {
		Name     string `json:"Name"`
		Password string `json:"Password"`
	}

	teamInfos := make([]*teamInfo, 0, len(f))

	for _, team := range f {
		password := randstr.String(16)

		// TODO add transaction
		_, err := db.Teams.Create(ctx.Request().Context(), db.CreateTeamOptions{
			Name:     team.Name,
			Password: password,
			Logo:     team.Logo,
		})
		if err != nil {
			if err == db.ErrTeamAlreadyExists {
				return ctx.Error(40000, l.T("team.exist", team.Name))
			}
			log.Error("Failed to create new team: %v", err)
			return ctx.ServerError()
		}

		teamInfos = append(teamInfos, &teamInfo{
			Name:     team.Name,
			Password: password,
		})
	}

	return ctx.Success(teamInfos)
}

// Update updates the team with the given options.
func (*TeamHandler) Update(ctx context.Context, f form.UpdateTeam) error {
	// Check the team exist or not.
	team, err := db.Teams.GetByID(ctx.Request().Context(), f.ID)
	if err != nil {
		log.Error("Failed to get team by ID: %v", err)
		return ctx.ServerError()
	}

	newTeam, err := db.Teams.GetByName(ctx.Request().Context(), f.Name)
	if err != nil {
		log.Error("Failed to get team by name: %v", err)
		return ctx.ServerError()
	}

	if team.ID != newTeam.ID {
		// TODO i18n
		return ctx.Error(40000, "Team name %q repeat.")
	}

	err = db.Teams.Update(ctx.Request().Context(), f.ID, db.UpdateTeamOptions{
		Name: f.Name,
		Logo: f.Logo,
	})
	if err != nil {
		log.Error("Failed to update team: %v", err)
		return ctx.ServerError()
	}
	return ctx.Success("")
}

// Delete deletes the team with the given ID.
func (*TeamHandler) Delete(ctx context.Context) error {
	id := uint(ctx.QueryInt("id"))

	// Check the team exist or not.
	team, err := db.Teams.GetByID(ctx.Request().Context(), id)
	if err != nil {
		log.Error("Failed to get team by ID: %v", err)
		return ctx.ServerError()
	}

	err = db.Teams.DeleteByID(ctx.Request().Context(), team.ID)
	if err != nil {
		log.Error("Failed to delete team: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success("")
}

// ResetPassword resets team password with the given id.
func (*TeamHandler) ResetPassword(ctx context.Context) error {
	id := uint(ctx.QueryInt("id"))

	// Check the team exist or not.
	team, err := db.Teams.GetByID(ctx.Request().Context(), id)
	if err != nil {
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
