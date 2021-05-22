// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package db

import (
	"testing"

	"gorm.io/gorm"

	"github.com/vidar-team/Cardinal/internal/dbutil"
)

func newTestDB(t *testing.T) (*gorm.DB, func(...string) error) {
	return dbutil.NewTestDB(t, allTables...)
}
