// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package form

type NewGameBox []struct {
	ChallengeID uint   `validate:"required"`
	TeamID      uint   `validate:"required"`
	Address     string `validate:"required"`
	Description string
	SSHPort     uint
	SSHUser     string
	SSHPassword string
}

type UpdateGameBox struct {
	ID          uint   `validate:"required"`
	Address     string `validate:"required"`
	Description string
	SSHPort     uint
	SSHUser     string
	SSHPassword string
}
