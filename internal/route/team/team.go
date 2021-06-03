// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package team

import (
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
)

func Authenticator(ctx context.Context) error {
	// TODO: get the team id from authorization header.
	team, err := db.Teams.GetByID(ctx.Request().Context(), 1)
	if err != nil {
		if err == db.ErrTeamNotExists {
			return ctx.Error(40300, "")
		}

		log.Error("Failed to get team: %v", err)
		return ctx.ServerError()
	}

	ctx.Map(team)
	return nil
}

// GetInfo returns the current logged in team's information.
func GetInfo(ctx context.Context, team *db.Team) error {
	return ctx.Success(map[string]interface{}{
		"Name":  team.Name,
		"Logo":  team.Logo,
		"Score": team.Score,
		"Rank":  team.Rank,
		"Token": team.Token,
	})
}
