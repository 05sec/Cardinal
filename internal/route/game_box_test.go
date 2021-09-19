// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flamego/flamego"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"

	"github.com/vidar-team/Cardinal/internal/form"
)

func TestGameBox(t *testing.T) {
	router, managerToken, cleanup := NewTestRoute(t)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, router *flamego.Flame, managerToken string)
	}{
		{"List", testListGameBoxes},
		{"New", testNewGameBox},
		{"Update", testUpdateGameBox},
		{"ResetAll", testResetAllGameBox},
		{"SSHTest", testSSHTestGameBox},
		{"RefreshFlag", testRefreshFlagGameBox},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("game_boxes", "challenges")
				if err != nil {
					t.Fatal(err)
				}
			})

			tc.test(t, router, managerToken)
		})
	}
}

func testListGameBoxes(t *testing.T, router *flamego.Flame, managerToken string) {
	// Empty game box list.
	req, err := http.NewRequest(http.MethodGet, "/api/manager/gameBoxes", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{"error":0,"data":[]}`
	assert.JSONEq(t, want, w.Body.String())

	// Create the base challenge.
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{FLAG}} > /flag",
	})

	// Create two game boxes.
	//createGameBox(t, managerToken, router, form.NewGameBox{{
	//	ChallengeID: 0,
	//	TeamID:      0,
	//	Address:     "",
	//	Description: "",
	//	SSHPort:     0,
	//	SSHUser:     "",
	//	SSHPassword: "",
	//}})

	// TODO Get the two game boxes.
}

func testNewGameBox(t *testing.T, router *flamego.Flame, managerToken string) {

}

func testUpdateGameBox(t *testing.T, router *flamego.Flame, managerToken string) {

}

func testResetAllGameBox(t *testing.T, router *flamego.Flame, managerToken string) {

}

func testSSHTestGameBox(t *testing.T, router *flamego.Flame, managerToken string) {

}

func testRefreshFlagGameBox(t *testing.T, router *flamego.Flame, managerToken string) {

}

func createGameBox(t *testing.T, managerToken string, router *flamego.Flame, f form.NewGameBox) {
	bodyBytes, err := jsoniter.Marshal(f)
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, "/api/manager/gameBox", bytes.NewBuffer(bodyBytes))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
