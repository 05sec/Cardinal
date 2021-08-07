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

func TestBulletins(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)
	store := NewBulletinsStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *bulletins)
	}{
		{"Create", testBulletinsCreate},
		{"Get", testBulletinsGet},
		{"GetByID", testBulletinsGetByID},
		{"Update", testBulletinsUpdate},
		{"DeleteByID", testBulletinsDeleteByID},
		{"DeleteAll", testBulletinsDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("bulletins")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), store.(*bulletins))
		})
	}
}

func testBulletinsCreate(t *testing.T, ctx context.Context, db *bulletins) {
	id, err := db.Create(ctx, CreateBulletinOptions{
		Title: "Welcome to D^3CTF",
		Body:  "Hey CTFer! Welcome to D^3CTF!",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)
}

func testBulletinsGet(t *testing.T, ctx context.Context, db *bulletins) {
	// Get empty bulletins lists.
	got, err := db.Get(ctx)
	assert.Nil(t, err)
	want := []*Bulletin{}
	assert.Equal(t, want, got)

	id, err := db.Create(ctx, CreateBulletinOptions{
		Title: "Welcome to D^3CTF",
		Body:  "Hey CTFer! Welcome to D^3CTF!",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	id, err = db.Create(ctx, CreateBulletinOptions{
		Title: "Web1 Updated",
		Body:  "Web1 Updated. Please check the hint.",
	})
	assert.Equal(t, uint(2), id)
	assert.Nil(t, err)

	got, err = db.Get(ctx)
	assert.Nil(t, err)

	for _, bulletin := range got {
		bulletin.CreatedAt = time.Time{}
		bulletin.UpdatedAt = time.Time{}
		bulletin.DeletedAt = gorm.DeletedAt{}
	}

	want = []*Bulletin{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Title: "Welcome to D^3CTF",
			Body:  "Hey CTFer! Welcome to D^3CTF!",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Title: "Web1 Updated",
			Body:  "Web1 Updated. Please check the hint.",
		},
	}
	assert.Equal(t, want, got)
}

func testBulletinsGetByID(t *testing.T, ctx context.Context, db *bulletins) {
	id, err := db.Create(ctx, CreateBulletinOptions{
		Title: "Welcome to D^3CTF",
		Body:  "Hey CTFer! Welcome to D^3CTF!",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)
	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}
	got.DeletedAt = gorm.DeletedAt{}

	want := &Bulletin{
		Model: gorm.Model{
			ID: 1,
		},
		Title: "Welcome to D^3CTF",
		Body:  "Hey CTFer! Welcome to D^3CTF!",
	}
	assert.Equal(t, want, got)

	// Get not exist bulletin.
	got, err = db.GetByID(ctx, 2)
	assert.Equal(t, ErrBulletinNotExists, err)
	want = (*Bulletin)(nil)
	assert.Equal(t, want, got)
}

func testBulletinsUpdate(t *testing.T, ctx context.Context, db *bulletins) {
	id, err := db.Create(ctx, CreateBulletinOptions{
		Title: "Welcome to D^3CTF",
		Body:  "Hey CTFer! Welcome to D^3CTF!",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	err = db.Update(ctx, 1, UpdateBulletinOptions{
		Title: "Welcome to D^3CTF!!!!",
		Body:  "Hey CTFer! Nice to meet you here!",
	})
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)
	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}
	got.DeletedAt = gorm.DeletedAt{}

	want := &Bulletin{
		Model: gorm.Model{
			ID: 1,
		},
		Title: "Welcome to D^3CTF!!!!",
		Body:  "Hey CTFer! Nice to meet you here!",
	}
	assert.Equal(t, want, got)
}

func testBulletinsDeleteByID(t *testing.T, ctx context.Context, db *bulletins) {
	id, err := db.Create(ctx, CreateBulletinOptions{
		Title: "Welcome to D^3CTF",
		Body:  "Hey CTFer! Welcome to D^3CTF!",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	err = db.DeleteByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Equal(t, ErrBulletinNotExists, err)
	want := (*Bulletin)(nil)
	assert.Equal(t, want, got)
}

func testBulletinsDeleteAll(t *testing.T, ctx context.Context, db *bulletins) {
	id, err := db.Create(ctx, CreateBulletinOptions{
		Title: "Welcome to D^3CTF",
		Body:  "Hey CTFer! Welcome to D^3CTF!",
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	id, err = db.Create(ctx, CreateBulletinOptions{
		Title: "Web1 Updated",
		Body:  "Web1 Updated. Please check the hint.",
	})
	assert.Equal(t, uint(2), id)
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	got, err := db.Get(ctx)
	assert.Nil(t, err)
	want := []*Bulletin{}
	assert.Equal(t, want, got)
}
