// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package team

import (
	"github.com/flamego/session"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
)

const teamIDSessionKey = "TeamID"

func Authenticator(ctx context.Context, session session.Session) error {
	teamID, ok := session.Get(teamIDSessionKey).(uint)
	if !ok {
		return ctx.Error(40300, "")
	}

	team, err := db.Teams.GetByID(ctx.Request().Context(), teamID)
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

func Login(ctx context.Context, session session.Session) error {
	// TODO: get the name and password from binding.
	team, err := db.Teams.Authenticate(ctx.Request().Context(), "", "")
	if err == db.ErrBadCredentials {
		return ctx.Error(40300, "bad credentials")
	} else if err != nil {
		log.Error("Failed to authenticate team: %v", err)
		return ctx.Error(50000, "")
	}

	session.Set(teamIDSessionKey, team.ID)
	return ctx.Success(session.ID())
}

func Logout(ctx context.Context, session session.Session) error {
	session.Flush()
	return ctx.Success("")
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
