// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package form

type NewChallenge struct {
	Title            string  `binding:"Required;MaxSize(255)"`
	BaseScore        float64 `binding:"Required;Range(0,10000)"`
	AutoRenewFlag    bool
	RenewFlagCommand string
}

type UpdateChallenge struct {
	ID               uint    `binding:"Required"`
	Title            string  `binding:"Required;MaxSize(255)"`
	BaseScore        float64 `binding:"Required;Range(0,10000)"`
	AutoRenewFlag    bool
	RenewFlagCommand string
}
