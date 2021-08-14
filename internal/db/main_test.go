// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"testing"

	"gorm.io/gorm"

	"github.com/vidar-team/Cardinal/internal/dbutil"
)

// newTestDB returns a test database instance with the cleanup function.
func newTestDB(t *testing.T) (*gorm.DB, func(...string) error) {
	return dbutil.NewTestDB(t, AllTables...)
}
