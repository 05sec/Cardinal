// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Cardinal-Platform/testify/assert"
	"github.com/flamego/flamego"
	jsoniter "github.com/json-iterator/go"

	"github.com/vidar-team/Cardinal/internal/form"
)

func TestGameBox(t *testing.T) {
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
		Title:            "Web1",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: "echo {{FLAG}} > /flag",
	})
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:     "Web2",
		BaseScore: 1500,
	})

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
				err := cleanup("game_boxes")
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

	// Create four game boxes of two challenges for two teams.
	createGameBox(t, managerToken, router, form.NewGameBox{
		{
			ChallengeID: 1,
			TeamID:      1,
			Address:     "192.168.1.1",
			Description: "Web1 for Vidar",
			SSHPort:     22,
			SSHUser:     "root",
			SSHPassword: "passw0rd",
		},
		{
			ChallengeID: 2,
			TeamID:      1,
			Address:     "192.168.2.1",
			Description: "Web2 for Vidar",
			SSHPort:     22,
			SSHUser:     "root",
			SSHPassword: "s3cret",
		},
		{
			ChallengeID: 1,
			TeamID:      2,
			Address:     "192.168.1.2",
			Description: "Web1 for E99p1ant",
			SSHPort:     22,
			SSHUser:     "root",
			SSHPassword: "passw0rd",
		},
		{
			ChallengeID: 2,
			TeamID:      2,
			Address:     "192.168.2.2",
			Description: "Web2 for E99p1ant",
			SSHPort:     22,
			SSHUser:     "root",
			SSHPassword: "s3cret",
		},
	})

	// Get the four game boxes.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/gameBoxes", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want = `{
    "data": [
        {
            "Address": "192.168.1.1",
            "Score": 1000,
            "DeletedAt": null,
            "InternalSSHPort": "22",
            "ChallengeID": 1,
            "TeamID": 1,
            "Challenge": {
                "ID": 1,
                "DeletedAt": null,
                "Title": "Web1",
                "BaseScore": 1000,
                "AutoRenewFlag": true,
                "RenewFlagCommand": "echo {{FLAG}} \u003e /flag"
            },
            "InternalSSHUser": "root",
            "InternalSSHPassword": "passw0rd",
            "Status": "up",
            "ID": 1,
            "Team": {
                "Token": "mocked_randstr_hex",
                "Logo": "https://vidar.club/logo.png",
                "Rank": 1,
                "ID": 1,
                "DeletedAt": null,
                "Name": "Vidar",
                "Score": 0
            },
            "Description": "Web1 for Vidar",
            "Visible": false
        },
        {
            "DeletedAt": null,
            "Description": "Web2 for Vidar",
            "InternalSSHPort": "22",
            "InternalSSHPassword": "s3cret",
            "Visible": false,
            "Score": 1500,
            "ID": 2,
            "ChallengeID": 2,
            "Challenge": {
                "Title": "Web2",
                "BaseScore": 1500,
                "AutoRenewFlag": false,
                "RenewFlagCommand": "",
                "ID": 2
            },
            "InternalSSHUser": "root",
            "TeamID": 1,
            "Team": {
                "ID": 1,
                "Name": "Vidar",
                "Score": 0,
                "Rank": 1,
                "Token": "mocked_randstr_hex",
                "DeletedAt": null,
                "Logo": "https://vidar.club/logo.png"
            },
            "Address": "192.168.2.1",
            "Status": "up"
        },
        {
            "Address": "192.168.1.2",
            "InternalSSHPassword": "passw0rd",
            "Visible": false,
            "Team": {
                "ID": 2,
                "Logo": "https://github.red/logo.png",
                "Score": 0,
                "Token": "mocked_randstr_hex",
                "DeletedAt": null,
                "Name": "E99p1ant",
                "Rank": 1
            },
            "Challenge": {
                "AutoRenewFlag": true,
                "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                "ID": 1,
                "DeletedAt": null,
                "Title": "Web1",
                "BaseScore": 1000
            },
            "ChallengeID": 1,
            "InternalSSHUser": "root",
            "Score": 1000,
            "Status": "up",
            "TeamID": 2,
            "Description": "Web1 for E99p1ant",
            "InternalSSHPort": "22",
            "ID": 3,
            "DeletedAt": null
        },
        {
            "DeletedAt": null,
            "Team": {
                "Logo": "https://github.red/logo.png",
                "Score": 0,
                "ID": 2,
                "DeletedAt": null,
                "Token": "mocked_randstr_hex",
                "Name": "E99p1ant",
                "Rank": 1
            },
            "TeamID": 2,
            "Challenge": {
                "ID": 2,
                "DeletedAt": null,
                "Title": "Web2",
                "BaseScore": 1500,
                "AutoRenewFlag": false,
                "RenewFlagCommand": ""
            },
            "Description": "Web2 for E99p1ant",
            "InternalSSHPassword": "s3cret",
            "Visible": false,
            "Score": 1500,
            "InternalSSHPort": "22",
            "ID": 4,
            "ChallengeID": 2,
            "Address": "192.168.2.2",
            "InternalSSHUser": "root",
            "Status": "up"
        }
    ],
    "error": 0
}`
	assert.JSONPartialEq(t, want, w.Body.String())
}

func testNewGameBox(t *testing.T, router *flamego.Flame, managerToken string) {
	for _, tc := range []struct {
		name            string
		body            string
		wantStatusCode  int
		wantGameBoxList string
	}{
		{
			name: "normal",
			body: `[{
				"ChallengeID": 1,
				"TeamID": 1,
				"Description": "Web1 for Vidar",
				"SSHUser": "root",
				"Address": "192.168.1.1",
				"SSHPort": 22,
				"SSHPassword": "passw0rd",
				"Score": 1000,
				"DeletedAt": null,
				"InternalSSHPort": "22"
			}]`,
			wantStatusCode: http.StatusOK,
			wantGameBoxList: `{
    "error": 0,
    "data": [
        {
            "Team": {
                "Name": "Vidar",
                "Score": 0,
                "Rank": 1,
                "Token": "mocked_randstr_hex",
                "ID": 1,
                "Logo": "https://vidar.club/logo.png"
            },
            "Challenge": {
                "Title": "Web1",
                "BaseScore": 1000,
                "AutoRenewFlag": true,
                "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                "ID": 1
            },
            "InternalSSHPort": "22",
            "InternalSSHPassword": "passw0rd",
            "Status": "up",
            "ID": 1,
            "ChallengeID": 1,
            "Address": "192.168.1.1",
            "TeamID": 1,
            "Description": "Web1 for Vidar",
            "InternalSSHUser": "root",
            "Visible": false,
            "Score": 1000
        }
    ]
}`,
		},
		{
			name: "duplicate game box",
			body: `[{
				"ChallengeID": 1,
				"TeamID": 1,
				"Description": "Web1 for Vidar",
				"SSHUser": "admin",
				"Address": "192.168.1.2",
				"SSHPort": 22,
				"SSHPassword": "s3cret",
				"Score": 1000,
				"DeletedAt": null,
				"InternalSSHPort": "22"
			}]`,
			wantStatusCode: http.StatusBadRequest,
			wantGameBoxList: `{
    "error": 0,
    "data": [
        {
            "Team": {
                "Name": "Vidar",
                "Score": 0,
                "Rank": 1,
                "Token": "mocked_randstr_hex",
                "ID": 1,
                "Logo": "https://vidar.club/logo.png"
            },
            "Challenge": {
                "Title": "Web1",
                "BaseScore": 1000,
                "AutoRenewFlag": true,
                "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                "ID": 1
            },
            "InternalSSHPort": "22",
            "InternalSSHPassword": "passw0rd",
            "Status": "up",
            "ID": 1,
            "ChallengeID": 1,
            "Address": "192.168.1.1",
            "TeamID": 1,
            "Description": "Web1 for Vidar",
            "InternalSSHUser": "root",
            "Visible": false,
            "Score": 1000
        }
    ]
}`,
		},
		{
			name: "create multiple game box",
			body: `[{
				"ChallengeID": 2,
				"TeamID": 1,
				"Description": "Web2 for Vidar",
				"SSHUser": "root",
				"Address": "192.168.2.1",
				"SSHPort": 22,
				"SSHPassword": "passw0rd",
				"Score": 1000,
				"InternalSSHPort": "22"
			},
			{
				"ChallengeID": 2,
				"TeamID": 2,
				"Description": "Web2 for E99p1ant",
				"SSHUser": "root",
				"Address": "192.168.2.2",
				"SSHPort": 22,
				"SSHPassword": "passw0rd",
				"Score": 1000,
				"InternalSSHPort": "22"
			}]`,
			wantStatusCode: http.StatusOK,
			wantGameBoxList: `{
    "error": 0,
    "data": [
        {
            "TeamID": 1,
            "Status": "up",
            "Team": {
                "ID": 1,
                "Name": "Vidar",
                "Logo": "https://vidar.club/logo.png",
                "Score": 0,
                "Rank": 1,
                "Token": "mocked_randstr_hex"
            },
            "ChallengeID": 1,
            "Challenge": {
                "AutoRenewFlag": true,
                "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                "ID": 1,
                "Title": "Web1",
                "BaseScore": 1000
            },
            "InternalSSHPort": "22",
            "Score": 1000,
            "ID": 1,
            "Description": "Web1 for Vidar",
            "Visible": false,
            "Address": "192.168.1.1",
            "InternalSSHUser": "root",
            "InternalSSHPassword": "passw0rd"
        },
        {
            "ID": 2,
            "TeamID": 1,
            "Description": "Web2 for Vidar",
            "Visible": false,
            "ChallengeID": 2,
            "Address": "192.168.2.1",
            "InternalSSHPassword": "passw0rd",
            "Status": "up",
            "Team": {
                "Name": "Vidar",
                "Logo": "https://vidar.club/logo.png",
                "Rank": 1,
                "Score": 0,
                "Token": "mocked_randstr_hex",
                "ID": 1
            },
            "Score": 1500,
            "Challenge": {
                "Title": "Web2",
                "BaseScore": 1500,
                "AutoRenewFlag": false,
                "RenewFlagCommand": "",
                "ID": 2
            },
            "InternalSSHPort": "22",
            "InternalSSHUser": "root"
        },
        {
            "Description": "Web2 for E99p1ant",
            "InternalSSHUser": "root",
            "TeamID": 2,
            "Team": {
                "Name": "E99p1ant",
                "Score": 0,
                "ID": 2,
                "Logo": "https://github.red/logo.png",
                "Rank": 1,
                "Token": "mocked_randstr_hex"
            },
            "ChallengeID": 2,
            "Challenge": {
                "ID": 2,
                "Title": "Web2",
                "BaseScore": 1500,
                "AutoRenewFlag": false,
                "RenewFlagCommand": ""
            },
            "Visible": false,
            "Score": 1500,
            "ID": 3,
            "InternalSSHPassword": "passw0rd",
            "Address": "192.168.2.2",
            "InternalSSHPort": "22",
            "Status": "up"
        }
    ]
}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/manager/gameBox", strings.NewReader(tc.body))
			assert.Nil(t, err)
			req.Header.Set("Authorization", managerToken)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			gotStatusCode := w.Code
			assert.Equal(t, tc.wantStatusCode, gotStatusCode)

			gotGameBoxList := getGameBoxes(t, managerToken, router)
			assert.JSONPartialEq(t, tc.wantGameBoxList, gotGameBoxList)
		})
	}
}

func testUpdateGameBox(t *testing.T, router *flamego.Flame, managerToken string) {
	createGameBox(t, managerToken, router, form.NewGameBox{
		{
			ChallengeID: 1,
			TeamID:      1,
			Address:     "192.168.1.1",
			Description: "Web1 for Vidar",
			SSHPort:     22,
			SSHUser:     "root",
			SSHPassword: "passw0rd",
		},
	})

	for _, tc := range []struct {
		name            string
		body            string
		wantStatusCode  int
		wantGameBoxList string
	}{
		{
			name: "normal",
			body: `{
				"ID": 1,
				"Address": "192.168.11.1",
				"Description": "Web1 for Vidar-Team",
				"SSHPort": 2222,
				"SSHUser": "root",
				"SSHPassword": "s3cret"
			}`,
			wantStatusCode: http.StatusOK,
			wantGameBoxList: `{
    "error": 0,
    "data": [
        {
            "TeamID": 1,
            "InternalSSHPort": "2222",
            "Address": "192.168.11.1",
            "Visible": false,
            "Score": 1000,
            "ID": 1,
            "Team": {
                "Logo": "https://vidar.club/logo.png",
                "Score": 0,
                "Rank": 1,
                "Token": "mocked_randstr_hex",
                "ID": 1,
                "Name": "Vidar"
            },
            "Challenge": {
                "Title": "Web1",
                "BaseScore": 1000,
                "AutoRenewFlag": true,
                "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                "ID": 1
            },
            "InternalSSHPassword": "s3cret",
            "ChallengeID": 1,
            "Description": "Web1 for Vidar-Team",
            "InternalSSHUser": "root",
            "Status": "up"
        }
    ]
}`,
		},
		{
			name: "game box not found",
			body: `{
				"ID": 2,
				"Address": "192.168.11.1",
				"Description": "Web1 for Vidar-Team",
				"SSHPort": 2222,
				"SSHUser": "root",
				"SSHPassword": "s3cret"
			}`,
			wantStatusCode: http.StatusNotFound,
			wantGameBoxList: `{
    "error": 0,
    "data": [
        {
            "TeamID": 1,
            "InternalSSHPort": "2222",
            "Address": "192.168.11.1",
            "Visible": false,
            "Score": 1000,
            "ID": 1,
            "Team": {
                "Logo": "https://vidar.club/logo.png",
                "Score": 0,
                "Rank": 1,
                "Token": "mocked_randstr_hex",
                "ID": 1,
                "Name": "Vidar"
            },
            "Challenge": {
                "Title": "Web1",
                "BaseScore": 1000,
                "AutoRenewFlag": true,
                "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                "ID": 1
            },
            "InternalSSHPassword": "s3cret",
            "ChallengeID": 1,
            "Description": "Web1 for Vidar-Team",
            "InternalSSHUser": "root",
            "Status": "up"
        }
    ]
}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPut, "/api/manager/gameBox", strings.NewReader(tc.body))
			assert.Nil(t, err)
			req.Header.Set("Authorization", managerToken)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			gotStatusCode := w.Code
			assert.Equal(t, tc.wantStatusCode, gotStatusCode)

			gotGameBoxList := getGameBoxes(t, managerToken, router)
			assert.JSONPartialEq(t, tc.wantGameBoxList, gotGameBoxList)
		})
	}
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

func getGameBoxes(t *testing.T, managerToken string, router *flamego.Flame) string {
	req, err := http.NewRequest(http.MethodGet, "/api/manager/gameBoxes", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	return w.Body.String()
}
