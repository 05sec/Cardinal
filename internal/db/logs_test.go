// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestLogs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)
	store := NewLogsStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *logs)
	}{
		{"Create", testLogsCreate},
		{"Get", testLogsGet},
		{"DeleteAll", testLogsDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("logs")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), store.(*logs))
		})
	}
}

func testLogsCreate(t *testing.T, ctx context.Context, db *logs) {
	err := db.Create(ctx, CreateLogOptions{
		Level: LogLevelNormal,
		Type:  LogTypeSystem,
		Body:  "Welcome to Cardinal!",
	})
	assert.Nil(t, err)

	err = db.Create(ctx, CreateLogOptions{
		Level: 4,
		Type:  LogTypeSystem,
		Body:  "Welcome to Cardinal!",
	})
	assert.Equal(t, ErrBadLogLevel, err)

	err = db.Create(ctx, CreateLogOptions{
		Level: LogLevelNormal,
		Type:  "unexpected_type",
		Body:  "Welcome to Cardinal!",
	})
	assert.Equal(t, ErrBadLogType, err)
}

func testLogsGet(t *testing.T, ctx context.Context, db *logs) {
	err := db.Create(ctx, CreateLogOptions{
		Level: LogLevelNormal,
		Type:  LogTypeSystem,
		Body:  "Welcome to Cardinal!",
	})
	assert.Nil(t, err)

	err = db.Create(ctx, CreateLogOptions{
		Level: LogLevelImportant,
		Type:  LogTypeSystem,
		Body:  "Please update your password!",
	})
	assert.Nil(t, err)

	got, err := db.Get(ctx)
	assert.Nil(t, err)

	for _, log := range got {
		log.CreatedAt = time.Time{}
		log.UpdatedAt = time.Time{}
	}

	want := []*Log{
		{
			Model: gorm.Model{
				ID: 2,
			},
			Level: LogLevelImportant,
			Type:  LogTypeSystem,
			Body:  "Please update your password!",
		},
		{
			Model: gorm.Model{
				ID: 1,
			},
			Level: LogLevelNormal,
			Type:  LogTypeSystem,
			Body:  "Welcome to Cardinal!",
		},
	}
	assert.Equal(t, want, got)
}

func testLogsDeleteAll(t *testing.T, ctx context.Context, db *logs) {
	err := db.Create(ctx, CreateLogOptions{
		Level: LogLevelNormal,
		Type:  LogTypeSystem,
		Body:  "Welcome to Cardinal!",
	})
	assert.Nil(t, err)

	err = db.Create(ctx, CreateLogOptions{
		Level: LogLevelImportant,
		Type:  LogTypeSystem,
		Body:  "Please update your password!",
	})
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	got, err := db.Get(ctx)
	assert.Nil(t, err)
	want := []*Log{}
	assert.Equal(t, want, got)
}
