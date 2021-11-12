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

func TestActions(t *testing.T) {
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

	gameBoxStore := NewGameBoxesStore(db)
	_, err = gameBoxStore.Create(ctx, CreateGameBoxOptions{
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

	_, err = gameBoxStore.Create(ctx, CreateGameBoxOptions{
		TeamID:      2,
		ChallengeID: 1,
		IPAddress:   "192.168.2.1",
		Port:        8080,
		Description: "Web1 For E99p1ant",
		InternalSSH: SSHConfig{
			Port:     22,
			User:     "root",
			Password: "s3crets",
		},
	})
	assert.Nil(t, err)

	actionsStore := NewActionsStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *actions)
	}{
		{"Create", testActionsCreate},
		{"Get", testActionsGet},
		{"SetScore", testActionsSetScore},
		{"CountScore", testActionsCountScore},
		{"GetEmptyScore", testActionsGetEmptyScore},
		{"Delete", testActionsDelete},
		{"DeleteAll", testActionsDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("actions")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), actionsStore.(*actions))
		})
	}
}

func testActionsCreate(t *testing.T, ctx context.Context, db *actions) {
	got, err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeBeenAttack,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &Action{
		Model: gorm.Model{
			ID: 1,
		},
		Type:           ActionTypeBeenAttack,
		TeamID:         1,
		ChallengeID:    1,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
		Score:          0,
	}
	assert.Equal(t, want, got)

	_, err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeBeenAttack,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Equal(t, ErrDuplicateAction, err)

	// Game box not found.
	_, err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeBeenAttack,
		GameBoxID:      3,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Equal(t, ErrGameBoxNotExists, err)
}

func testActionsGet(t *testing.T, ctx context.Context, db *actions) {
	_, err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeBeenAttack,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeAttack,
		GameBoxID:      1,
		AttackerTeamID: 1, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		GameBoxID:      2,
		AttackerTeamID: 2, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	// Get actions by team ID.
	got, err := db.Get(ctx, GetActionOptions{
		TeamID: 1,
	})
	assert.Nil(t, err)

	for _, action := range got {
		action.CreatedAt = time.Time{}
		action.UpdatedAt = time.Time{}
	}

	want := []*Action{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Type:           ActionTypeBeenAttack,
			TeamID:         1,
			ChallengeID:    1,
			GameBoxID:      1,
			AttackerTeamID: 2,
			Round:          1,
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Type:           ActionTypeAttack,
			TeamID:         1,
			ChallengeID:    1,
			GameBoxID:      1,
			AttackerTeamID: 0,
			Round:          1,
		},
	}
	assert.Equal(t, want, got)

	// Get actions by action ID.
	got, err = db.Get(ctx, GetActionOptions{
		ActionID: 1,
	})
	assert.Nil(t, err)

	for _, action := range got {
		action.CreatedAt = time.Time{}
		action.UpdatedAt = time.Time{}
	}

	want = []*Action{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Type:           ActionTypeBeenAttack,
			TeamID:         1,
			ChallengeID:    1,
			GameBoxID:      1,
			AttackerTeamID: 2,
			Round:          1,
		},
	}
	assert.Equal(t, want, got)
}

func testActionsSetScore(t *testing.T, ctx context.Context, db *actions) {
	_, err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeAttack,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)

	err = db.SetScore(ctx, SetActionScoreOptions{
		Round:     1,
		GameBoxID: 1,
		Score:     50,
	})
	assert.Nil(t, err)
	action, err := db.Get(ctx, GetActionOptions{
		GameBoxID: 1,
		Round:     1,
	})
	assert.Nil(t, err)
	assert.NotZero(t, action)
	assert.Equal(t, float64(50), action[0].Score)

	// Replace score
	err = db.SetScore(ctx, SetActionScoreOptions{
		Round:     1,
		GameBoxID: 1,
		Score:     60,
		Replace:   true,
	})
	assert.Nil(t, err)
	action, err = db.Get(ctx, GetActionOptions{
		GameBoxID: 1,
		Round:     1,
	})
	assert.Nil(t, err)
	assert.NotZero(t, action)
	assert.Equal(t, float64(60), action[0].Score)

	err = db.SetScore(ctx, SetActionScoreOptions{
		Round:     2,
		GameBoxID: 1,
		Score:     50,
	})
	assert.Nil(t, err)

	// Invalid score, set a negative number to a AttackAction.
	err = db.SetScore(ctx, SetActionScoreOptions{
		Round:     1,
		GameBoxID: 1,
		Score:     -50,
	})
	assert.Equal(t, ErrActionScoreInvalid, err)
}

func testActionsCountScore(t *testing.T, ctx context.Context, db *actions) {
	action, err := db.Create(ctx, CreateActionOptions{Type: ActionTypeBeenAttack, GameBoxID: 1, AttackerTeamID: 2, Round: 1})
	assert.Nil(t, err)
	err = db.SetScore(ctx, SetActionScoreOptions{ActionID: action.ID, Score: -50, Replace: false})
	assert.Nil(t, err)

	action, err = db.Create(ctx, CreateActionOptions{Type: ActionTypeServiceOnline, GameBoxID: 1, AttackerTeamID: 2, Round: 1})
	assert.Nil(t, err)
	err = db.SetScore(ctx, SetActionScoreOptions{ActionID: action.ID, Score: 70, Replace: false})
	assert.Nil(t, err)

	action, err = db.Create(ctx, CreateActionOptions{Type: ActionTypeServiceOnline, GameBoxID: 2, AttackerTeamID: 2, Round: 1})
	assert.Nil(t, err)
	err = db.SetScore(ctx, SetActionScoreOptions{ActionID: action.ID, Score: 50, Replace: false})
	assert.Nil(t, err)

	got, err := db.CountScore(ctx, CountActionScoreOptions{
		GameBoxID: 1,
		Round:     1,
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(20), got)

	got, err = db.CountScore(ctx, CountActionScoreOptions{
		Round: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, float64(70), got)
}

func testActionsGetEmptyScore(t *testing.T, ctx context.Context, db *actions) {
	_, err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeBeenAttack,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)

	got, err := db.GetEmptyScore(ctx, 1, ActionTypeBeenAttack)
	assert.Nil(t, err)

	for _, score := range got {
		score.CreatedAt = time.Time{}
		score.UpdatedAt = time.Time{}
	}

	want := []*Action{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Type:           ActionTypeBeenAttack,
			TeamID:         1,
			ChallengeID:    1,
			GameBoxID:      1,
			AttackerTeamID: 2,
			Round:          1,
			Score:          0,
		},
	}
	assert.Equal(t, want, got)

	err = db.SetScore(ctx, SetActionScoreOptions{
		Round:     1,
		GameBoxID: 1,
		Score:     -60,
		Replace:   true,
	})
	assert.Nil(t, err)

	got, err = db.GetEmptyScore(ctx, 1, ActionTypeBeenAttack)
	assert.Nil(t, err)
	want = []*Action{}
	assert.Equal(t, want, got)
}

func testActionsDelete(t *testing.T, ctx context.Context, db *actions) {
	_, err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeBeenAttack,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		GameBoxID:      1,
		AttackerTeamID: 1, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		GameBoxID:      2,
		AttackerTeamID: 2, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	// Delete by ID.
	err = db.Delete(ctx, DeleteActionOptions{
		ActionID: 2,
	})
	assert.Nil(t, err)
	got, err := db.Get(ctx, GetActionOptions{})
	assert.Nil(t, err)

	for _, action := range got {
		action.CreatedAt = time.Time{}
		action.UpdatedAt = time.Time{}
	}

	want := []*Action{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Type:           ActionTypeBeenAttack,
			TeamID:         1,
			ChallengeID:    1,
			GameBoxID:      1,
			AttackerTeamID: 2,
			Round:          1,
			Score:          0,
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			Type:           ActionTypeCheckDown,
			TeamID:         2,
			ChallengeID:    1,
			GameBoxID:      2,
			AttackerTeamID: 0,
			Round:          1,
			Score:          0,
		},
	}
	assert.Equal(t, want, got)

	// Delete nothing.
	err = db.Delete(ctx, DeleteActionOptions{
		ChallengeID: 1,
		Round:       2,
	})
	assert.Nil(t, err)
	got, err = db.Get(ctx, GetActionOptions{})
	assert.Nil(t, err)

	for _, action := range got {
		action.CreatedAt = time.Time{}
		action.UpdatedAt = time.Time{}
	}

	want = []*Action{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Type:           ActionTypeBeenAttack,
			TeamID:         1,
			ChallengeID:    1,
			GameBoxID:      1,
			AttackerTeamID: 2,
			Round:          1,
			Score:          0,
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			Type:           ActionTypeCheckDown,
			TeamID:         2,
			ChallengeID:    1,
			GameBoxID:      2,
			AttackerTeamID: 0,
			Round:          1,
			Score:          0,
		},
	}
	assert.Equal(t, want, got)

	// Delete by challenge ID.
	err = db.Delete(ctx, DeleteActionOptions{
		ChallengeID: 1,
	})
	assert.Nil(t, err)
	_, err = db.Get(ctx, GetActionOptions{})
	assert.Nil(t, err)
}

func testActionsDeleteAll(t *testing.T, ctx context.Context, db *actions) {
	_, err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeAttack,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		GameBoxID:      1,
		AttackerTeamID: 1, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		GameBoxID:      2,
		AttackerTeamID: 2, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	_, err = db.Get(ctx, GetActionOptions{})
	assert.Nil(t, err)
}
