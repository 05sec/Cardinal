// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package manager

import (
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
)

// GetTeams returns all the teams.
func (*Handler) GetTeams(ctx context.Context) error {
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

func (*Handler) NewTeams(ctx context.Context) error {
	return nil
}

func (*Handler) UpdateTeam(ctx context.Context) error {
	return nil
}

func (*Handler) DeleteTeam(ctx context.Context) error {
	return nil
}

func (*Handler) ResetTeamPassword(ctx context.Context) error {
	return nil
}
