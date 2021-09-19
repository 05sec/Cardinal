// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package form

type NewChallenge struct {
	Title            string  `validate:"required,lt=255"`
	BaseScore        float64 `validate:"required,gte=0,lte=10000"`
	AutoRenewFlag    bool
	RenewFlagCommand string
}

type UpdateChallenge struct {
	ID               uint    `validate:"required"`
	Title            string  `validate:"required,lt=255"`
	BaseScore        float64 `validate:"required,gte=0,lte=10000"`
	AutoRenewFlag    bool
	RenewFlagCommand string
}

type SetChallengeVisible struct {
	ID      uint `binding:"Required"`
	Visible bool
}
