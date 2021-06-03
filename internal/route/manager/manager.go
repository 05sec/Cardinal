// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package manager

import (
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
)

func Authenticator(ctx context.Context) error {
	// TODO: get the manager id from authorization header.
	manager, err := db.Managers.GetByID(ctx.Request().Context(), 1)
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
