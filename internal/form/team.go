// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package form

type TeamLogin struct {
	Name     string `binding:"Required;MaxSize(255)"`
	Password string `binding:"Required;MaxSize(255)"`
}

type NewTeam []struct {
	Name string `binding:"Required;MaxSize(255)"`
	Logo string `binding:"Required;MaxSize(255)"`
}
