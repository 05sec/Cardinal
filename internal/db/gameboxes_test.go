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

func TestGameBoxes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)

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

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *gameboxes)
	}{
		{"Create", testGameBoxesCreate},
		{"Get", testGameBoxesGet},
		{"GetByID", testGameBoxesGetByID},
		{"Update", testGameBoxesUpdate},
		{"SetScore", testGameBoxesSetScore},
		{"SetVisible", testGameBoxesSetVisible},
		{"SetStatus", testGameBoxesSetStatus},
		{"DeleteByID", testGameBoxesDeleteByID},
		{"DeleteAll", testGameBoxesDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("game_boxes")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), gameBoxesStore.(*gameboxes))
		})
	}
}

func testGameBoxesCreate(t *testing.T, ctx context.Context, db *gameboxes) {
	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, ErrGameBoxAlreadyExists, err)
}

func testGameBoxesGet(t *testing.T, ctx context.Context, db *gameboxes) {
	// Get empty game box list.
	gameboxes, err := db.Get(ctx)
	assert.Nil(t, err)
	assert.Empty(t, gameboxes)

	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	id, err = db.Create(ctx, CreateGameBoxOptions{
		TeamID:      2,
		ChallengeID: 1,
		Address:     "192.168.2.1",
		Description: "Web1 For E99p1ant",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "s3crets",
		},
	})
	assert.Equal(t, uint(2), id)
	assert.Nil(t, err)

	got, err := db.Get(ctx)
	assert.Nil(t, err)

	for _, gameBox := range got {
		gameBox.CreatedAt = time.Time{}
		gameBox.UpdatedAt = time.Time{}
	}

	teamsStore := NewTeamsStore(db.DB)
	team1, err := teamsStore.GetByID(ctx, 1)
	assert.Nil(t, err)
	team2, err := teamsStore.GetByID(ctx, 2)
	assert.Nil(t, err)

	challengesStore := NewChallengesStore(db.DB)
	challenge, err := challengesStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	want := []*GameBox{
		{
			Model: gorm.Model{
				ID: 1,
			},
			TeamID:              1,
			Team:                team1,
			ChallengeID:         1,
			Challenge:           challenge,
			Address:             "192.168.1.1",
			Description:         "Web1 For Vidar",
			InternalSSHPort:     "22",
			InternalSSHUser:     "root",
			InternalSSHPassword: "passw0rd",
			Visible:             false,
			Score:               1000,
			Status:              GameBoxStatusUp,
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			TeamID:              2,
			Team:                team2,
			ChallengeID:         1,
			Challenge:           challenge,
			Address:             "192.168.2.1",
			Description:         "Web1 For E99p1ant",
			InternalSSHPort:     "22",
			InternalSSHUser:     "root",
			InternalSSHPassword: "s3crets",
			Visible:             false,
			Score:               1000,
			Status:              GameBoxStatusUp,
		},
	}

	assert.Equal(t, want, got)
}

func testGameBoxesGetByID(t *testing.T, ctx context.Context, db *gameboxes) {
	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	teamsStore := NewTeamsStore(db.DB)
	team, err := teamsStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	challengesStore := NewChallengesStore(db.DB)
	challenge, err := challengesStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &GameBox{
		Model: gorm.Model{
			ID: 1,
		},
		TeamID:              1,
		Team:                team,
		ChallengeID:         1,
		Challenge:           challenge,
		Address:             "192.168.1.1",
		Description:         "Web1 For Vidar",
		InternalSSHPort:     "22",
		InternalSSHUser:     "root",
		InternalSSHPassword: "passw0rd",
		Visible:             false,
		Score:               1000,
		Status:              GameBoxStatusUp,
	}
	assert.Equal(t, want, got)

	_, err = db.GetByID(ctx, 2)
	assert.Equal(t, ErrGameBoxNotExists, err)
}

func testGameBoxesUpdate(t *testing.T, ctx context.Context, db *gameboxes) {
	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	err = db.Update(ctx, 1, UpdateGameBoxOptions{
		Address:     "192.168.1.11",
		Description: "This is the Web1, have fun!",
		InternalSSH: SSHConfig{
			Port:     2222,
			User:     "r00t",
			Password: "s3cret",
		},
	})
	assert.Nil(t, err)

	teamsStore := NewTeamsStore(db.DB)
	team, err := teamsStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	challengesStore := NewChallengesStore(db.DB)
	challenge, err := challengesStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &GameBox{
		Model: gorm.Model{
			ID: 1,
		},
		TeamID:              1,
		Team:                team,
		ChallengeID:         1,
		Challenge:           challenge,
		Address:             "192.168.1.11",
		Description:         "This is the Web1, have fun!",
		InternalSSHPort:     "2222",
		InternalSSHUser:     "r00t",
		InternalSSHPassword: "s3cret",
		Visible:             false,
		Score:               1000,
		Status:              GameBoxStatusUp,
	}
	assert.Equal(t, want, got)
}

func testGameBoxesSetScore(t *testing.T, ctx context.Context, db *gameboxes) {
	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	err = db.SetScore(ctx, 1, 1800)
	assert.Nil(t, err)

	teamsStore := NewTeamsStore(db.DB)
	team, err := teamsStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	challengesStore := NewChallengesStore(db.DB)
	challenge, err := challengesStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &GameBox{
		Model: gorm.Model{
			ID: 1,
		},
		TeamID:              1,
		Team:                team,
		ChallengeID:         1,
		Challenge:           challenge,
		Address:             "192.168.1.1",
		Description:         "Web1 For Vidar",
		InternalSSHPort:     "22",
		InternalSSHUser:     "root",
		InternalSSHPassword: "passw0rd",
		Visible:             false,
		Score:               1800,
		Status:              GameBoxStatusUp,
	}
	assert.Equal(t, want, got)
}

func testGameBoxesSetVisible(t *testing.T, ctx context.Context, db *gameboxes) {
	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	// Set true.
	err = db.SetVisible(ctx, 1, true)
	assert.Nil(t, err)

	teamsStore := NewTeamsStore(db.DB)
	team, err := teamsStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	challengesStore := NewChallengesStore(db.DB)
	challenge, err := challengesStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &GameBox{
		Model: gorm.Model{
			ID: 1,
		},
		TeamID:              1,
		Team:                team,
		ChallengeID:         1,
		Challenge:           challenge,
		Address:             "192.168.1.1",
		Description:         "Web1 For Vidar",
		InternalSSHPort:     "22",
		InternalSSHUser:     "root",
		InternalSSHPassword: "passw0rd",
		Visible:             true,
		Score:               1000,
		Status:              GameBoxStatusUp,
	}
	assert.Equal(t, want, got)

	// Set false.
	err = db.SetVisible(ctx, 1, false)
	assert.Nil(t, err)

	got, err = db.GetByID(ctx, 1)
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want = &GameBox{
		Model: gorm.Model{
			ID: 1,
		},
		TeamID:              1,
		Team:                team,
		ChallengeID:         1,
		Challenge:           challenge,
		Address:             "192.168.1.1",
		Description:         "Web1 For Vidar",
		InternalSSHPort:     "22",
		InternalSSHUser:     "root",
		InternalSSHPassword: "passw0rd",
		Visible:             false,
		Score:               1000,
		Status:              GameBoxStatusUp,
	}
	assert.Equal(t, want, got)
}

func testGameBoxesSetStatus(t *testing.T, ctx context.Context, db *gameboxes) {
	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	err = db.SetStatus(ctx, 1, GameBoxStatusCaptured)
	assert.Nil(t, err)

	teamsStore := NewTeamsStore(db.DB)
	team, err := teamsStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	challengesStore := NewChallengesStore(db.DB)
	challenge, err := challengesStore.GetByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &GameBox{
		Model: gorm.Model{
			ID: 1,
		},
		TeamID:              1,
		Team:                team,
		ChallengeID:         1,
		Challenge:           challenge,
		Address:             "192.168.1.1",
		Description:         "Web1 For Vidar",
		InternalSSHPort:     "22",
		InternalSSHUser:     "root",
		InternalSSHPassword: "passw0rd",
		Visible:             false,
		Score:               1000,
		Status:              GameBoxStatusCaptured,
	}
	assert.Equal(t, want, got)

	// Set unexpected game box status.
	err = db.SetStatus(ctx, 1, "unexpected")
	assert.Equal(t, ErrBadGameBoxsStatus, err)
}

func testGameBoxesDeleteByID(t *testing.T, ctx context.Context, db *gameboxes) {
	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	err = db.DeleteByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Equal(t, ErrGameBoxNotExists, err)
	want := (*GameBox)(nil)
	assert.Equal(t, want, got)
}

func testGameBoxesDeleteAll(t *testing.T, ctx context.Context, db *gameboxes) {
	id, err := db.Create(ctx, CreateGameBoxOptions{
		TeamID:      1,
		ChallengeID: 1,
		Address:     "192.168.1.1",
		Description: "Web1 For Vidar",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "passw0rd",
		},
	})
	assert.Equal(t, uint(1), id)
	assert.Nil(t, err)

	id, err = db.Create(ctx, CreateGameBoxOptions{
		TeamID:      2,
		ChallengeID: 1,
		Address:     "192.168.2.1",
		Description: "Web1 For E99p1ant",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "s3crets",
		},
	})
	assert.Equal(t, uint(2), id)
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	got, err := db.Get(ctx)
	assert.Nil(t, err)
	want := []*GameBox{}
	assert.Equal(t, want, got)
}
