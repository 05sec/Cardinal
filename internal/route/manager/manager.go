// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package manager

import (
	"github.com/flamego/session"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
)

// Handler is the manager request handler.
type Handler struct{}

// NewHandler creates and returns a new manager Handler.
func NewHandler() *Handler {
	return &Handler{}
}

const managerIDSessionKey = "TeamID"

func (*Handler) Authenticator(ctx context.Context, session session.Session) error {
	managerID, ok := session.Get(managerIDSessionKey).(uint)
	if !ok {
		return ctx.Error(40300, "")
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
