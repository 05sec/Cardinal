// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRanks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)
	ranksStore := NewRanksStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *ranks)
	}{
		{"List", testRanksList},
		{"VisibleChallengeTitle", testRanksVisibleChallengeTitle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("teams", "challenges", "game_boxes")
				if err != nil {
					t.Fatal(err)
				}
			})

			ctx := context.Background()
			// Create three teams.
			teamsStore := NewTeamsStore(db)
			_, err := teamsStore.Create(ctx, CreateTeamOptions{Name: "Vidar"})
			assert.Nil(t, err)
			_, err = teamsStore.Create(ctx, CreateTeamOptions{Name: "E99p1ant"})
			assert.Nil(t, err)
			_, err = teamsStore.Create(ctx, CreateTeamOptions{Name: "Cosmos"})
			assert.Nil(t, err)

			// Create two challenges.
			challengesStore := NewChallengesStore(db)
			_, err = challengesStore.Create(ctx, CreateChallengeOptions{Title: "Web1", BaseScore: 1000})
			assert.Nil(t, err)
			_, err = challengesStore.Create(ctx, CreateChallengeOptions{Title: "Pwn1", BaseScore: 1000})
			assert.Nil(t, err)

			// Create game boxes for each team and challenge.
			gameBoxesStore := NewGameBoxesStore(db)
			_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{TeamID: 1, ChallengeID: 1, Address: "192.168.1.1", Description: "Web1 For Vidar"})
			assert.Nil(t, err)
			_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{TeamID: 1, ChallengeID: 2, Address: "192.168.2.1", Description: "Pwn1 For Vidar"})
			assert.Nil(t, err)
			_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{TeamID: 2, ChallengeID: 1, Address: "192.168.1.2", Description: "Web1 For E99p1ant"})
			assert.Nil(t, err)
			_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{TeamID: 2, ChallengeID: 2, Address: "192.168.2.2", Description: "Pwn1 For E99p1ant"})
			assert.Nil(t, err)
			_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{TeamID: 3, ChallengeID: 1, Address: "192.168.1.3", Description: "Web1 For Cosmos"})
			assert.Nil(t, err)
			_, err = gameBoxesStore.Create(ctx, CreateGameBoxOptions{TeamID: 3, ChallengeID: 2, Address: "192.168.2.3", Description: "Pwn1 For Cosmos"})
			assert.Nil(t, err)

			tc.test(t, context.Background(), ranksStore.(*ranks))
		})
	}
}

func testRanksList(t *testing.T, ctx context.Context, db *ranks) {
	gameBoxesStore := NewGameBoxesStore(db.DB)
	for i := uint(1); i <= 6; i++ {
		err := gameBoxesStore.SetVisible(ctx, i, true)
		assert.Nil(t, err)
	}

	// Cosmos	1800 1000
	// Vidar	1000 700
	// E99p1ant 700  500
	err := gameBoxesStore.SetScore(ctx, 1, 1000)
	assert.Nil(t, err)
	err = gameBoxesStore.SetScore(ctx, 2, 700)
	assert.Nil(t, err)
	err = gameBoxesStore.SetScore(ctx, 3, 700)
	assert.Nil(t, err)
	err = gameBoxesStore.SetScore(ctx, 4, 500)
	assert.Nil(t, err)
	err = gameBoxesStore.SetScore(ctx, 5, 1800)
	assert.Nil(t, err)
	err = gameBoxesStore.SetScore(ctx, 6, 1000)
	assert.Nil(t, err)

	scoreStore := NewScoresStore(db.DB)
	err = scoreStore.RefreshTeamScore(ctx)
	assert.Nil(t, err)

	got, err := db.List(ctx)
	assert.Nil(t, err)

	want := []*RankItem{
		{
			TeamID:   3,
			TeamName: "Cosmos",
			TeamLogo: "",
			Rank:     1,
			Score:    2800,
			GameBoxes: []*GameBoxInfo{
				{ChallengeID: 1, Score: 1800},
				{ChallengeID: 2, Score: 1000},
			},
		},
		{
			TeamID:   1,
			TeamName: "Vidar",
			Rank:     2,
			Score:    1700,
			GameBoxes: []*GameBoxInfo{
				{ChallengeID: 1, Score: 1000},
				{ChallengeID: 2, Score: 700},
			},
		},
		{
			TeamID:   2,
			TeamName: "E99p1ant",
			Rank:     3,
			Score:    1200,
			GameBoxes: []*GameBoxInfo{
				{ChallengeID: 1, Score: 700},
				{ChallengeID: 2, Score: 500},
			},
		},
	}
	assert.Equal(t, want, got)
}

func testRanksVisibleChallengeTitle(t *testing.T, ctx context.Context, db *ranks) {
	// All the game boxes are invisible, so the ranking list title is nil.
	got, err := db.VisibleChallengeTitle(ctx)
	assert.Nil(t, err)
	want := []string{}
	assert.Equal(t, want, got)

	gameBoxesStore := NewGameBoxesStore(db.DB)
	for i := uint(1); i <= 6; i++ {
		err := gameBoxesStore.SetVisible(ctx, i, true)
		assert.Nil(t, err)
	}

	got, err = db.VisibleChallengeTitle(ctx)
	assert.Nil(t, err)
	want = []string{"Web1", "Pwn1"}
	assert.Equal(t, want, got)
}
