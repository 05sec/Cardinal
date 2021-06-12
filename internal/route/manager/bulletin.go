// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package manager

import (
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/form"
)

// GetBulletins returns all the bulletins.
func (*Handler) GetBulletins(ctx context.Context) error {
	bulletins, err := db.Bulletins.Get(ctx.Request().Context())
	if err != nil {
		log.Error("Failed to get bulletins list: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success(bulletins)
}

// NewBulletin creates a new bulletin with the given options.
func (*Handler) NewBulletin(ctx context.Context, f form.NewBulletin) error {
	_, err := db.Bulletins.Create(ctx.Request().Context(), db.CreateBulletinOptions{
		Title: f.Title,
		Body:  f.Body,
	})
	if err != nil {
		log.Error("Failed to create new bulletin: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success("")
}

// UpdateBulletin updates the bulletin with the given options.
func (*Handler) UpdateBulletin(ctx context.Context, f form.UpdateBulletin) error {
	// Check the bulletin exists or not.
	bulletin, err := db.Bulletins.GetByID(ctx.Request().Context(), f.ID)
	if err != nil {
		if err == db.ErrBulletinNotExists {
			return ctx.Error(40000, "Bulletin dose not exist.")
		}
		log.Error("Failed to get bulletin: %v", err)
		return ctx.ServerError()
	}

	err = db.Bulletins.Update(ctx.Request().Context(), bulletin.ID, db.UpdateBulletinOptions{
		Title: f.Title,
		Body:  f.Body,
	})
	if err != nil {
		log.Error("Failed to update bulletin: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success("")
}

// DeleteBulletin deletes the bulletin with the given id.
func (*Handler) DeleteBulletin(ctx context.Context) error {
	id := uint(ctx.QueryInt("id"))

	// Check the bulletin exists or not.
	bulletin, err := db.Bulletins.GetByID(ctx.Request().Context(), id)
	if err != nil {
		if err == db.ErrBulletinNotExists {
			return ctx.Error(40000, "Bulletin dose not exist.")
		}
		log.Error("Failed to get bulletin: %v", err)
		return ctx.ServerError()
	}

	err = db.Bulletins.DeleteByID(ctx.Request().Context(), bulletin.ID)
	if err != nil {
		log.Error("Failed to delete bulletin: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success("")
}
