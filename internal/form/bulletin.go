// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package form

type NewBulletin struct {
	Title string `binding:"Required;MaxSize(255)"`
	Body  string `binding:"Required;MaxSize(1000)"`
}

type UpdateBulletin struct {
	ID    uint   `binding:"Required"`
	Title string `binding:"Required;MaxSize(255)"`
	Body  string `binding:"Required;MaxSize(1000)"`
}
