// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
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
		// TODO get the info from main.go
		"Version": "",
		"Commit":  "",
	})
}

func (*GeneralHandler) Init(c context.Context) error {
	// TODO: Get value from config file.
	return c.Success(map[string]interface{}{
		"Title":    "",
		"Language": "",
	})
}

func (*GeneralHandler) NotFound(c context.Context) error {
	// TODO: i18n support.
	return c.Error(40400, "not found")
}
