// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package form

type ManagerLogin struct {
	Name     string `validate:"required,lt=255"`
	Password string `validate:"required,lt=255"`
}
