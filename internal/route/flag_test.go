// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Cardinal-Platform/testify/assert"
	"github.com/flamego/flamego"

	"github.com/vidar-team/Cardinal/internal/clock"
	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/form"
)

func TestFlag(t *testing.T) {
	router, managerToken, cleanup := NewTestRoute(t)

	// Create two teams.
	createTeam(t, managerToken, router, form.NewTeam{
		{
			Name: "Vidar",
			Logo: "https://vidar.club/logo.png",
		},
		{
			Name: "E99p1ant",
			Logo: "https://github.red/logo.png",
		},
	})

	// Create two basic challenges.
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:     "Web1",
		BaseScore: 1000,
	})
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:     "Web2",
		BaseScore: 1500,
	})

	// Create four game boxes.
	createGameBox(t, managerToken, router, form.NewGameBox{
		{ChallengeID: 1, TeamID: 1, IPAddress: "192.168.1.1", Port: 80, Description: "Web1 For Vidar"},
		{ChallengeID: 1, TeamID: 2, IPAddress: "192.168.1.2", Port: 8080, Description: "Web1 For E99p1ant"},
		{ChallengeID: 2, TeamID: 1, IPAddress: "192.168.2.1", Port: 80, Description: "Web2 For Vidar"},
		{ChallengeID: 2, TeamID: 2, IPAddress: "192.168.2.2", Port: 8080, Description: "Web2 For E99p1ant"},
	})

	conf.Game.FlagPrefix = "d3ctf{"
	conf.Game.FlagSuffix = "}"

	// Mock total round.
	totalRound := clock.T.TotalRound
	t.Cleanup(func() {
		clock.T.TotalRound = totalRound
	})
	clock.T.TotalRound = 2

	for _, tc := range []struct {
		name string
		test func(t *testing.T, router *flamego.Flame, managerToken string)
	}{
		{"Get", testFlagGet},
		{"BatchCreate", testFlagBatchCreate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("flags")
				if err != nil {
					t.Fatal(err)
				}
			})

			tc.test(t, router, managerToken)
		})
	}
}

func testFlagGet(t *testing.T, router *flamego.Flame, managerToken string) {
	// Empty flags.
	req, err := http.NewRequest(http.MethodGet, "/api/manager/flags", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{"error":0,"data": {"Count": 0, "List": []}}`
	assert.JSONEq(t, want, w.Body.String())

	// Create flags.
	createFlag(t, router, managerToken)
	req, err = http.NewRequest(http.MethodGet, "/api/manager/flags", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	want = `
{
    "data": {
		"Count": 8,
		"List": [
			{
				"ChallengeID": 1,
				"GameBoxID": 1,
				"ID": 1,
				"Round": 1,
				"TeamID": 1,
				"Value": "d3ctf{7046a8da1dbbb0b70d9843f1e19e1eee4679f727}"
			},
			{
				"ChallengeID": 1,
				"GameBoxID": 2,
				"ID": 2,
				"Round": 1,
				"TeamID": 2,
				"Value": "d3ctf{782049c46936c0e01897ca35f11611b3878b8990}"
			},
			{
				"ChallengeID": 2,
				"GameBoxID": 3,
				"ID": 3,
				"Round": 1,
				"TeamID": 1,
				"Value": "d3ctf{15b15829651d2551b0904040dfb3ebbe5519e5e7}"
			},
			{
				"ChallengeID": 2,
				"GameBoxID": 4,
				"ID": 4,
				"Round": 1,
				"TeamID": 2,
				"Value": "d3ctf{241f46eecf5d15cd0571cb0f6b6acc3de7d1b277}"
			},
			{
				"ChallengeID": 1,
				"GameBoxID": 1,
				"ID": 5,
				"Round": 2,
				"TeamID": 1,
				"Value": "d3ctf{e11756ef65f34c46782fead3babd0726624f3d2d}"
			},
			{
				"ChallengeID": 1,
				"GameBoxID": 2,
				"ID": 6,
				"Round": 2,
				"TeamID": 2,
				"Value": "d3ctf{40091ccbbc5e4f516250b9b8125c3ebfa7e20515}"
			},
			{
				"ChallengeID": 2,
				"GameBoxID": 3,
				"ID": 7,
				"Round": 2,
				"TeamID": 1,
				"Value": "d3ctf{f7941b09ee3ad7b078052958dd4b55e62acb9da7}"
			},
			{
				"ChallengeID": 2,
				"GameBoxID": 4,
				"ID": 8,
				"Round": 2,
				"TeamID": 2,
				"Value": "d3ctf{3263048c9b68ff1f7759df13269fdf250ac84f66}"
			}
    	]
	},
    "error": 0
}`
	assert.JSONPartialEq(t, want, w.Body.String())
}

func testFlagBatchCreate(t *testing.T, router *flamego.Flame, managerToken string) {
	req, err := http.NewRequest(http.MethodGet, "/api/manager/flags", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func createFlag(t *testing.T, router *flamego.Flame, managerToken string) {
	req, err := http.NewRequest(http.MethodPost, "/api/manager/flags", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
