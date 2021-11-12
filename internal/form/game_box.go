// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package form

type NewGameBox []struct {
	ChallengeID         uint   `validate:"required,lt=255"`
	TeamID              uint   `validate:"required,lt=255"`
	IPAddress           string `validate:"required,lt=255"`
	Port                uint
	Description         string
	InternalSSHPort     uint
	InternalSSHUser     string
	InternalSSHPassword string
}

type UpdateGameBox struct {
	ID                  uint   `validate:"required,lt=255"`
	IPAddress           string `validate:"required,lt=255"`
	Port                uint
	Description         string
	InternalSSHPort     uint
	InternalSSHUser     string
	InternalSSHPassword string
}
