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

// ManagerHandler is the manager request handler.
type ManagerHandler struct{}

// NewManagerHandler creates and returns a new manager Handler.
func NewManagerHandler() *ManagerHandler {
	return &ManagerHandler{}
}

const managerIDSessionKey = "ManagerID"

func (*ManagerHandler) Authenticator(ctx context.Context, session session.Session) error {
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

func (*ManagerHandler) Login(ctx context.Context, session session.Session, f form.ManagerLogin) error {
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

func (*ManagerHandler) Logout(ctx context.Context, session session.Session) error {
	session.Flush()
	return ctx.Success("")
}
