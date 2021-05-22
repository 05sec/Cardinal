// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestChallenges(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)
	store := NewChallengesStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *challenges)
	}{
		{"Create", testChallengesCreate},
		{"Get", testChallengesGet},
		{"GetByID", testChallengesGetByID},
		{"Update", testChallengesUpdate},
		{"DeleteByID", testChallengesDeleteByID},
		{"DeleteAll", testChallengesDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("challenges")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), store.(*challenges))
		})
	}
}

func testChallengesCreate(t *testing.T, ctx context.Context, db *challenges) {
	id, err := db.Create(ctx, CreateChallengeOptions{
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{flag}} > /flag",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateChallengeOptions{
		Title:            "Web1",
		BaseScore:        1500,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{flag}} > /flag",
	})
	assert.Equal(t, ErrChallengeAlreadyExists, err)
}

func testChallengesGet(t *testing.T, ctx context.Context, db *challenges) {
	// Get empty challenge lists.
	got, err := db.Get(ctx)
	assert.Nil(t, err)
	want := []*Challenge{}
	assert.Equal(t, want, got)

	id, err := db.Create(ctx, CreateChallengeOptions{
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{flag}} > /flag",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	id, err = db.Create(ctx, CreateChallengeOptions{
		Title:            "Pwn1",
		BaseScore:        1500,
		AutoRenewFlag:    false,
		RenewFlagCommand: "",
	})
	assert.Equal(t, uint(2), id)
	assert.Nil(t, err)

	got, err = db.Get(ctx)
	assert.Nil(t, err)

	for _, challenge := range got {
		challenge.CreatedAt = time.Time{}
		challenge.UpdatedAt = time.Time{}
		challenge.DeletedAt = gorm.DeletedAt{}
	}

	want = []*Challenge{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Title:            "Web1",
			BaseScore:        1000,
			AutoRenewFlag:    true,
			RenewFlagCommand: "echo {{flag}} > /flag",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Title:            "Pwn1",
			BaseScore:        1500,
			AutoRenewFlag:    false,
			RenewFlagCommand: "",
		},
	}
	assert.Equal(t, want, got)
}

func testChallengesGetByID(t *testing.T, ctx context.Context, db *challenges) {
	id, err := db.Create(ctx, CreateChallengeOptions{
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{flag}} > /flag",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)
	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}
	got.DeletedAt = gorm.DeletedAt{}

	want := &Challenge{
		Model: gorm.Model{
			ID: 1,
		},
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{flag}} > /flag",
	}
	assert.Equal(t, want, got)

	// Get not exist challenge.
	got, err = db.GetByID(ctx, 2)
	assert.Equal(t, ErrChallengeNotExists, err)
	want = (*Challenge)(nil)
	assert.Equal(t, want, got)
}

func testChallengesUpdate(t *testing.T, ctx context.Context, db *challenges) {
	id, err := db.Create(ctx, CreateChallengeOptions{
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{flag}} > /flag",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	err = db.Update(ctx, 1, UpdateChallengeOptions{
		Title:            "Web2",
		BaseScore:        500,
		AutoRenewFlag:    false,
		RenewFlagCommand: "echo 'flag'",
	})
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)
	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}
	got.DeletedAt = gorm.DeletedAt{}

	want := &Challenge{
		Model: gorm.Model{
			ID: 1,
		},
		Title:            "Web2",
		BaseScore:        500,
		AutoRenewFlag:    false,
		RenewFlagCommand: "echo 'flag'",
	}
	assert.Equal(t, want, got)
}

func testChallengesDeleteByID(t *testing.T, ctx context.Context, db *challenges) {
	id, err := db.Create(ctx, CreateChallengeOptions{
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{flag}} > /flag",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	err = db.DeleteByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Equal(t, ErrChallengeNotExists, err)
	want := (*Challenge)(nil)
	assert.Equal(t, want, got)
}

func testChallengesDeleteAll(t *testing.T, ctx context.Context, db *challenges) {
	id, err := db.Create(ctx, CreateChallengeOptions{
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{flag}} > /flag",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	id, err = db.Create(ctx, CreateChallengeOptions{
		Title:            "Pwn1",
		BaseScore:        1500,
		AutoRenewFlag:    false,
		RenewFlagCommand: "",
	})
	assert.Equal(t, uint(2), id)
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	got, err := db.Get(ctx)
	assert.Nil(t, err)
	want := []*Challenge{}
	assert.Equal(t, want, got)
}
