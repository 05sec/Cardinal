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

func TestFlags(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)
	flagsStore := NewFlagsStore(db)

	ctx := context.Background()
	challengesStore := NewChallengesStore(db)
	_, err := challengesStore.Create(ctx, CreateChallengeOptions{
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{FLAG}} > /flag",
	})
	assert.Nil(t, err)
	_, err = challengesStore.Create(ctx, CreateChallengeOptions{
		Title:            "Pwn1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{FLAG}} > /flag",
	})
	assert.Nil(t, err)

	teamsStore := NewTeamsStore(db)
	_, err = teamsStore.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)
	_, err = teamsStore.Create(ctx, CreateTeamOptions{
		Name:     "E99p1ant",
		Password: "asdfgh",
		Logo:     "https://github.red/",
	})
	assert.Nil(t, err)

	gameBoxesStore := NewGameBoxesStore(db)
	_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		IPAddress:   "192.168.1.1",
		Port:        80,
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Nil(t, err)
	_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{
		TeamID:      2,
		ChallengeID: 1,
		IPAddress:   "192.168.2.1",
		Port:        80,
		Description: "Web1 For E99p1ant",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "s3crets",
		},
	})
	assert.Nil(t, err)
	_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 2,
		IPAddress:   "192.168.1.2",
		Port:        8080,
		Description: "Web2 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "t0ken",
		},
	})
	assert.Nil(t, err)
	_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{
		TeamID:      2,
		ChallengeID: 2,
		IPAddress:   "192.168.2.2",
		Port:        8080,
		Description: "Web2 For E99p1ant",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "auth",
		},
	})
	assert.Nil(t, err)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *flags)
	}{
		{"BatchCreate", testFlagsBatchCreate},
		{"Get", testFlagsGet},
		{"Count", testFlagsCount},
		{"Check", testFlagsCheck},
		{"DeleteAll", testFlagsDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("flags")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), flagsStore.(*flags))
		})
	}
}

func testFlagsBatchCreate(t *testing.T, ctx context.Context, db *flags) {
	err := db.BatchCreate(ctx, CreateFlagOptions{
		Flags: []FlagMetadata{
			{
				GameBoxID: 1,
				Round:     1,
				Value:     "d3ctf{c4ca4238a0b923820dcc509a6f75849b}",
			},
			{
				GameBoxID: 2,
				Round:     1,
				Value:     "d3ctf{c81e728d9d4c2f636f067f89cc14862c}",
			},
			{
				GameBoxID: 3,
				Round:     1,
				Value:     "d3ctf{eccbc87e4b5ce2fe28308fd9f2a7baf3}",
			},
			{
				GameBoxID: 4,
				Round:     1,
				Value:     "d3ctf{a87ff679a2f3e71d9181a67b7542122c}",
			},
		},
	})
	assert.Nil(t, err)

	// Upsert flag.
	err = db.BatchCreate(ctx, CreateFlagOptions{
		Flags: []FlagMetadata{
			{
				GameBoxID: 1,
				Round:     1,
				Value:     "d3ctf{upsert_it}",
			},
		},
	})
	assert.Nil(t, err)

	flags, _, err := db.Get(ctx, GetFlagOptions{
		GameBoxID: 1,
		Round:     1,
	})
	assert.Nil(t, err)
	assert.NotZero(t, flags)
	got := flags[0].Value
	want := "d3ctf{upsert_it}"
	assert.Equal(t, want, got)

	// New round
	err = db.BatchCreate(ctx, CreateFlagOptions{
		Flags: []FlagMetadata{
			{
				GameBoxID: 1,
				Round:     2,
				Value:     "d3ctf{81092a3cf05843109f022d3ff201ec6e}",
			},
			{
				GameBoxID: 2,
				Round:     2,
				Value:     "d3ctf{dc891cef907e0615aeb659018f43e186}",
			},
			{
				GameBoxID: 3,
				Round:     2,
				Value:     "d3ctf{8fff5617c428b42497a81ebffff4f92a}",
			},
			{
				GameBoxID: 4,
				Round:     2,
				Value:     "d3ctf{bf97cc7e6f6088aa2837338cbc61e95c}",
			},
		},
	})
	assert.Nil(t, err)

	// Game box not found.
	err = db.BatchCreate(ctx, CreateFlagOptions{
		Flags: []FlagMetadata{
			{
				GameBoxID: 5,
				Round:     1,
				Value:     "d3ctf{e4da3b7fbbce2345d7772b0674a318d5}",
			},
			{
				GameBoxID: 6,
				Round:     1,
				Value:     "d3ctf{1679091c5a880faf6fb5e6087eb1b2dc}",
			},
		},
	})
	assert.Equal(t, ErrGameBoxNotExists, err)
}

func testFlagsGet(t *testing.T, ctx context.Context, db *flags) {
	err := db.BatchCreate(ctx, CreateFlagOptions{
		Flags: []FlagMetadata{
			{
				GameBoxID: 1,
				Round:     1,
				Value:     "d3ctf{c4ca4238a0b923820dcc509a6f75849b}",
			},
			{
				GameBoxID: 2,
				Round:     1,
				Value:     "d3ctf{c81e728d9d4c2f636f067f89cc14862c}",
			},
			{
				GameBoxID: 3,
				Round:     1,
				Value:     "d3ctf{eccbc87e4b5ce2fe28308fd9f2a7baf3}",
			},
			{
				GameBoxID: 4,
				Round:     1,
				Value:     "d3ctf{a87ff679a2f3e71d9181a67b7542122c}",
			},
		},
	})
	assert.Nil(t, err)

	got, gotCount, err := db.Get(ctx, GetFlagOptions{})
	assert.Nil(t, err)

	for _, flag := range got {
		flag.CreatedAt = time.Time{}
		flag.UpdatedAt = time.Time{}
	}

	want := []*Flag{
		{
			Model: gorm.Model{
				ID: 1,
			},
			TeamID:      1,
			ChallengeID: 1,
			GameBoxID:   1,
			Round:       1,
			Value:       "d3ctf{c4ca4238a0b923820dcc509a6f75849b}",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			TeamID:      2,
			ChallengeID: 1,
			GameBoxID:   2,
			Round:       1,
			Value:       "d3ctf{c81e728d9d4c2f636f067f89cc14862c}",
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			TeamID:      1,
			ChallengeID: 2,
			GameBoxID:   3,
			Round:       1,
			Value:       "d3ctf{eccbc87e4b5ce2fe28308fd9f2a7baf3}",
		},
		{
			Model: gorm.Model{
				ID: 4,
			},
			TeamID:      2,
			ChallengeID: 2,
			GameBoxID:   4,
			Round:       1,
			Value:       "d3ctf{a87ff679a2f3e71d9181a67b7542122c}",
		},
	}
	assert.Equal(t, want, got)
	assert.Equal(t, int64(4), gotCount)

	got, gotCount, err = db.Get(ctx, GetFlagOptions{
		Page:     1,
		PageSize: 3,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(4), gotCount)

	for _, flag := range got {
		flag.CreatedAt = time.Time{}
		flag.UpdatedAt = time.Time{}
	}

	want = []*Flag{
		{
			Model: gorm.Model{
				ID: 1,
			},
			TeamID:      1,
			ChallengeID: 1,
			GameBoxID:   1,
			Round:       1,
			Value:       "d3ctf{c4ca4238a0b923820dcc509a6f75849b}",
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			TeamID:      2,
			ChallengeID: 1,
			GameBoxID:   2,
			Round:       1,
			Value:       "d3ctf{c81e728d9d4c2f636f067f89cc14862c}",
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			TeamID:      1,
			ChallengeID: 2,
			GameBoxID:   3,
			Round:       1,
			Value:       "d3ctf{eccbc87e4b5ce2fe28308fd9f2a7baf3}",
		},
	}
	assert.Equal(t, want, got)
}

func testFlagsCount(t *testing.T, ctx context.Context, db *flags) {
	err := db.BatchCreate(ctx, CreateFlagOptions{
		Flags: []FlagMetadata{
			{
				GameBoxID: 1,
				Round:     1,
				Value:     "d3ctf{c4ca4238a0b923820dcc509a6f75849b}",
			},
			{
				GameBoxID: 2,
				Round:     1,
				Value:     "d3ctf{c81e728d9d4c2f636f067f89cc14862c}",
			},
			{
				GameBoxID: 3,
				Round:     1,
				Value:     "d3ctf{eccbc87e4b5ce2fe28308fd9f2a7baf3}",
			},
			{
				GameBoxID: 4,
				Round:     1,
				Value:     "d3ctf{a87ff679a2f3e71d9181a67b7542122c}",
			},
		},
	})
	assert.Nil(t, err)

	got, err := db.Count(ctx, CountFlagOptions{})
	assert.Nil(t, err)
	assert.Equal(t, int64(4), got)

	got, err = db.Count(ctx, CountFlagOptions{
		GameBoxID: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), got)
}

func testFlagsCheck(t *testing.T, ctx context.Context, db *flags) {
	err := db.BatchCreate(ctx, CreateFlagOptions{
		Flags: []FlagMetadata{
			{
				GameBoxID: 1,
				Round:     1,
				Value:     "d3ctf{c4ca4238a0b923820dcc509a6f75849b}",
			},
			{
				GameBoxID: 2,
				Round:     1,
				Value:     "d3ctf{c81e728d9d4c2f636f067f89cc14862c}",
			},
			{
				GameBoxID: 3,
				Round:     1,
				Value:     "d3ctf{eccbc87e4b5ce2fe28308fd9f2a7baf3}",
			},
			{
				GameBoxID: 4,
				Round:     1,
				Value:     "d3ctf{a87ff679a2f3e71d9181a67b7542122c}",
			},
		},
	})
	assert.Nil(t, err)

	got, err := db.Check(ctx, "d3ctf{c81e728d9d4c2f636f067f89cc14862c}")
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &Flag{
		Model: gorm.Model{
			ID: 2,
		},
		TeamID:      2,
		ChallengeID: 1,
		GameBoxID:   2,
		Round:       1,
		Value:       "d3ctf{c81e728d9d4c2f636f067f89cc14862c}",
	}
	assert.Equal(t, want, got)

	_, err = db.Check(ctx, "d3ctf{not_found}")
	assert.Equal(t, ErrFlagNotExists, err)
}

func testFlagsDeleteAll(t *testing.T, ctx context.Context, db *flags) {
	err := db.BatchCreate(ctx, CreateFlagOptions{
		Flags: []FlagMetadata{
			{
				GameBoxID: 1,
				Round:     1,
				Value:     "d3ctf{c4ca4238a0b923820dcc509a6f75849b}",
			},
			{
				GameBoxID: 2,
				Round:     1,
				Value:     "d3ctf{c81e728d9d4c2f636f067f89cc14862c}",
			},
			{
				GameBoxID: 3,
				Round:     1,
				Value:     "d3ctf{eccbc87e4b5ce2fe28308fd9f2a7baf3}",
			},
			{
				GameBoxID: 4,
				Round:     1,
				Value:     "d3ctf{a87ff679a2f3e71d9181a67b7542122c}",
			},
		},
	})
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	got, gotCount, err := db.Get(ctx, GetFlagOptions{})
	assert.Nil(t, err)
	want := []*Flag{}
	assert.Equal(t, want, got)
	assert.Equal(t, int64(0), gotCount)
}
