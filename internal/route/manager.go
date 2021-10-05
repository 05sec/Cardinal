// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/rank"
)

// ManagerHandler is the manager request handler.
type ManagerHandler struct{}

// NewManagerHandler creates and returns a new manager Handler.
func NewManagerHandler() *ManagerHandler {
	return &ManagerHandler{}
}

func (*ManagerHandler) Panel() {

}

func (*ManagerHandler) Logs() {

}

func (*ManagerHandler) Rank(ctx context.Context) error {
	return ctx.Success(rank.ForManager())
}
