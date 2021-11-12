// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"github.com/flamego/session"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/form"
)

// AuthHandler is the authenticate request handler.
type AuthHandler struct{}

// NewAuthHandler creates and returns a new authenticate Handler.
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

const teamIDSessionKey = "TeamID"
const managerIDSessionKey = "ManagerID"

func (*AuthHandler) TeamAuthenticator(ctx context.Context, session session.Session) error {
	teamID, ok := session.Get(teamIDSessionKey).(uint)
	if !ok {
		return ctx.Error(40300, "team authenticate error")
	}

	team, err := db.Teams.GetByID(ctx.Request().Context(), teamID)
	if err != nil {
		if err == db.ErrTeamNotExists {
			return ctx.Error(40300, "")
		}

		log.Error("Failed to get team by ID: %v", err)
		return ctx.ServerError()
	}

	ctx.Map(team)
	return nil
}

func (*AuthHandler) TeamTokenAuthenticator(ctx context.Context) error {
	token := ctx.Query("token")
	team, err := db.Teams.GetByToken(ctx.Request().Context(), token)
	if err != nil {
		if err == db.ErrTeamNotExists {
			return ctx.Error(40300, "")
		}

		log.Error("Failed to get team by token: %v", err)
		return ctx.ServerError()
	}

	ctx.Map(team)
	return nil
}

func (*AuthHandler) TeamLogin(ctx context.Context, session session.Session, f form.TeamLogin) error {
	team, err := db.Teams.Authenticate(ctx.Request().Context(), f.Name, f.Password)
	if err == db.ErrBadCredentials {
		return ctx.Error(40300, "bad credentials")
	} else if err != nil {
		log.Error("Failed to authenticate team: %v", err)
		return ctx.Error(50000, "")
	}

	session.Set(teamIDSessionKey, team.ID)
	return ctx.Success(session.ID())
}

func (*AuthHandler) TeamLogout(ctx context.Context, session session.Session) error {
	session.Delete(teamIDSessionKey)
	return ctx.Success()
}

func (*AuthHandler) ManagerAuthenticator(ctx context.Context, session session.Session) error {
	managerID, ok := session.Get(managerIDSessionKey).(uint)
	if !ok {
		return ctx.Error(40300, "manager authenticate error")
	}

	manager, err := db.Managers.GetByID(ctx.Request().Context(), managerID)
	if err != nil {
		if err == db.ErrManagerNotExists {
			return ctx.Error(40300, "")
		}

		log.Error("Failed to get manager: %v", err)
		return ctx.ServerError()
	}

	ctx.Map(manager)
	return nil
}

func (*AuthHandler) ManagerLogin(ctx context.Context, session session.Session, f form.ManagerLogin) error {
	manager, err := db.Managers.Authenticate(ctx.Request().Context(), f.Name, f.Password)
	if err == db.ErrBadCredentials {
		return ctx.Error(40300, "bad credentials")
	} else if err != nil {
		log.Error("Failed to authenticate manager: %v", err)
		return ctx.Error(50000, "")
	}

	session.Set(managerIDSessionKey, manager.ID)
	return ctx.Success(session.ID())
}

func (*AuthHandler) ManagerLogout(ctx context.Context, session session.Session) error {
	session.Delete(managerIDSessionKey)
	return ctx.Success()
}
