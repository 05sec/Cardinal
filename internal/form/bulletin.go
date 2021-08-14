// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package form

type NewBulletin struct {
	Title string `validate:"required,lt=255"`
	Body  string `validate:"required,lt=1000"`
}

type UpdateBulletin struct {
	ID    uint   `validate:"required"`
	Title string `validate:"required,lt(255)"`
	Body  string `validate:"required,lt(1000)"`
}
