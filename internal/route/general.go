// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"time"

	"github.com/vidar-team/Cardinal/internal/clock"
	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/context"
)

// GeneralHandler is the general request handler.
type GeneralHandler struct{}

// NewGeneralHandler creates and returns a new GeneralHandler.
func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}

func (*GeneralHandler) Hello(c context.Context) error {
	return c.Success(map[string]interface{}{
		"Version":     conf.Version,
		"BuildTime":   conf.BuildTime,
		"BuildCommit": conf.BuildCommit,
	})
}

func (*GeneralHandler) Init(c context.Context) error {
	return c.Success(map[string]interface{}{
		"Name": conf.App.Name,
	})
}

func (*GeneralHandler) Time(c context.Context) error {
	return c.Success(map[string]interface{}{
		"CurrentTime":         time.Now().Unix(),
		"StartAt":             clock.T.StartAt.Unix(),
		"EndAt":               clock.T.EndAt.Unix(),
		"RoundDuration":       clock.T.RoundDuration.Seconds(),
		"CurrentRound":        clock.T.CurrentRound,
		"RoundRemainDuration": int(clock.T.RoundRemainDuration.Seconds()),
		"Status":              clock.T.Status,
		"TotalRound":          clock.T.TotalRound,
	})
}

func (*GeneralHandler) NotFound(c context.Context) error {
	// TODO: i18n support.
	return c.Error(40400, "not found")
}
