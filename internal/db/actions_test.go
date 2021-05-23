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

func TestActions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)
	store := NewActionsStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *actions)
	}{
		{"Create", testActionsCreate},
		{"Get", testActionsGet},
		{"DeleteAll", testActionsDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("actions")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), store.(*actions))
		})
	}
}

func testActionsCreate(t *testing.T, ctx context.Context, db *actions) {
	err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeAttack,
		TeamID:         1,
		ChallengeID:    1,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)
}

func testActionsGet(t *testing.T, ctx context.Context, db *actions) {
	err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeAttack,
		TeamID:         1,
		ChallengeID:    1,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)

	err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		TeamID:         1,
		ChallengeID:    1,
		GameBoxID:      1,
		AttackerTeamID: 1, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		TeamID:         2,
		ChallengeID:    1,
		GameBoxID:      2,
		AttackerTeamID: 2, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

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
			Type:           ActionTypeAttack,
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
			Type:           ActionTypeCheckDown,
			TeamID:         1,
			ChallengeID:    1,
			GameBoxID:      1,
			AttackerTeamID: 0,
			Round:          1,
		},
	}
	assert.Equal(t, want, got)
}

func testActionsDeleteAll(t *testing.T, ctx context.Context, db *actions) {
	err := db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeAttack,
		TeamID:         1,
		ChallengeID:    1,
		GameBoxID:      1,
		AttackerTeamID: 2,
		Round:          1,
	})
	assert.Nil(t, err)

	err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		TeamID:         1,
		ChallengeID:    1,
		GameBoxID:      1,
		AttackerTeamID: 1, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	err = db.Create(ctx, CreateActionOptions{
		Type:           ActionTypeCheckDown,
		TeamID:         2,
		ChallengeID:    1,
		GameBoxID:      2,
		AttackerTeamID: 2, // Will be set to 0.
		Round:          1,
	})
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	got, err := db.Get(ctx, GetActionOptions{})
	assert.Nil(t, err)
	want := []*Action{}
	assert.Equal(t, want, got)
}
