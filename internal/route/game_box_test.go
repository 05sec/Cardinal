// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"fmt"
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
	want := `{"error":0,"data": {"Data": [], "Count": 0}}`
	assert.JSONEq(t, want, w.Body.String())

	// Create four game boxes of two challenges for two teams.
	createGameBox(t, managerToken, router, form.NewGameBox{
		{
			ChallengeID:         1,
			TeamID:              1,
			IPAddress:           "192.168.1.1",
			Port:                80,
			Description:         "Web1 for Vidar",
			InternalSSHPort:     22,
			InternalSSHUser:     "root",
			InternalSSHPassword: "passw0rd",
		},
		{
			ChallengeID:         2,
			TeamID:              1,
			IPAddress:           "192.168.2.1",
			Port:                8080,
			Description:         "Web2 for Vidar",
			InternalSSHPort:     22,
			InternalSSHUser:     "root",
			InternalSSHPassword: "s3cret",
		},
		{
			ChallengeID:         1,
			TeamID:              2,
			IPAddress:           "192.168.1.2",
			Port:                80,
			Description:         "Web1 for E99p1ant",
			InternalSSHPort:     22,
			InternalSSHUser:     "root",
			InternalSSHPassword: "passw0rd",
		},
		{
			ChallengeID:         2,
			TeamID:              2,
			IPAddress:           "192.168.2.2",
			Port:                8080,
			Description:         "Web2 for E99p1ant",
			InternalSSHPort:     22,
			InternalSSHUser:     "root",
			InternalSSHPassword: "s3cret",
		},
	})

	// Get the four game boxes.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/gameBoxes", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	fmt.Println(w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
	want = `{
    "data": {
        "Count": 4,
        "Data": [
            {
                "Challenge": {
                    "AutoRenewFlag": true,
                    "BaseScore": 1000,
                    "ID": 1,
                    "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                    "Title": "Web1"
                },
                "ChallengeID": 1,
                "Description": "Web1 for Vidar",
                "ID": 1,
                "IPAddress": "192.168.1.1",
                "InternalSSHPassword": "passw0rd",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 80,
                "Score": 1000,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 1,
                    "Logo": "https://vidar.club/logo.png",
                    "Name": "Vidar",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 1,
                "Visible": false
            },
            {
                "Challenge": {
                    "AutoRenewFlag": false,
                    "BaseScore": 1500,
                    "ID": 2,
                    "RenewFlagCommand": "",
                    "Title": "Web2"
                },
                "ChallengeID": 2,
                "Description": "Web2 for Vidar",
                "ID": 2,
                "IPAddress": "192.168.2.1",
                "InternalSSHPassword": "s3cret",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 8080,
                "Score": 1500,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 1,
                    "Logo": "https://vidar.club/logo.png",
                    "Name": "Vidar",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 1,
                "Visible": false
            },
            {
                "Challenge": {
                    "AutoRenewFlag": true,
                    "BaseScore": 1000,
                    "ID": 1,
                    "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                    "Title": "Web1"
                },
                "ChallengeID": 1,
                "Description": "Web1 for E99p1ant",
                "ID": 3,
                "IPAddress": "192.168.1.2",
                "InternalSSHPassword": "passw0rd",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 80,
                "Score": 1000,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 2,
                    "Logo": "https://github.red/logo.png",
                    "Name": "E99p1ant",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 2,
                "Visible": false
            },
            {
                "Challenge": {
                    "AutoRenewFlag": false,
                    "BaseScore": 1500,
                    "ID": 2,
                    "RenewFlagCommand": "",
                    "Title": "Web2"
                },
                "ChallengeID": 2,
                "Description": "Web2 for E99p1ant",
                "ID": 4,
                "IPAddress": "192.168.2.2",
                "InternalSSHPassword": "s3cret",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 8080,
                "Score": 1500,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 2,
                    "Logo": "https://github.red/logo.png",
                    "Name": "E99p1ant",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 2,
                "Visible": false
            }
        ]
    },
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
			body: `
[
    {
        "ChallengeID": 1,
        "Description": "Web1 for Vidar",
        "IPAddress": "192.168.1.1",
        "Port": 80,
        "InternalSSHPassword": "passw0rd",
        "InternalSSHPort": 22,
        "InternalSSHUser": "root",
        "Score": 1000,
        "TeamID": 1
    }
]
`,
			wantStatusCode: http.StatusOK,
			wantGameBoxList: `
{
    "data": {
        "Count": 1,
        "Data": [
            {
                "Challenge": {
                    "AutoRenewFlag": true,
                    "BaseScore": 1000,
                    "ID": 1,
                    "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                    "Title": "Web1"
                },
                "ChallengeID": 1,
                "Description": "Web1 for Vidar",
                "ID": 1,
                "IPAddress": "192.168.1.1",
                "InternalSSHPassword": "passw0rd",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 80,
                "Score": 1000,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 1,
                    "Logo": "https://vidar.club/logo.png",
                    "Name": "Vidar",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 1,
                "Visible": false
            }
        ]
    },
    "error": 0
}
`,
		},
		{
			name: "duplicate game box",
			body: `
[
    {
        "ChallengeID": 1,
        "Description": "Web1 for Vidar",
        "IPAddress": "192.168.1.2",
        "Port": 8080,
        "InternalSSHPassword": "s3cret",
        "InternalSSHPort": 22,
        "InternalSSHUser": "admin",
        "Score": 1000,
        "TeamID": 1
    }
]
`,
			wantStatusCode: http.StatusBadRequest,
			wantGameBoxList: `
{
    "data": {
        "Count": 1,
        "Data": [
            {
                "Challenge": {
                    "AutoRenewFlag": true,
                    "BaseScore": 1000,
                    "ID": 1,
                    "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                    "Title": "Web1"
                },
                "ChallengeID": 1,
                "Description": "Web1 for Vidar",
                "ID": 1,
                "IPAddress": "192.168.1.1",
                "InternalSSHPassword": "passw0rd",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 80,
                "Score": 1000,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 1,
                    "Logo": "https://vidar.club/logo.png",
                    "Name": "Vidar",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 1,
                "Visible": false
            }
        ]
    },
    "error": 0
}
`,
		},
		{
			name: "create multiple game box",
			body: `
[
    {
        "ChallengeID": 2,
        "Description": "Web2 for Vidar",
        "IPAddress": "192.168.2.1",
        "Port": 8080,
        "InternalSSHPassword": "passw0rd",
        "InternalSSHPort": 22,
        "InternalSSHUser": "root",
        "Score": 1000,
        "TeamID": 1
    },
    {
        "ChallengeID": 2,
        "Description": "Web2 for E99p1ant",
        "IPAddress": "192.168.2.2",
        "Port": 8080,
        "InternalSSHPassword": "passw0rd",
        "InternalSSHPort": 22,
        "InternalSSHUser": "root",
        "Score": 1000,
        "TeamID": 2
    }
]
`,
			wantStatusCode: http.StatusOK,
			wantGameBoxList: `
{
    "data": {
        "Count": 3,
        "Data": [
            {
                "Challenge": {
                    "AutoRenewFlag": true,
                    "BaseScore": 1000,
                    "ID": 1,
                    "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                    "Title": "Web1"
                },
                "ChallengeID": 1,
                "Description": "Web1 for Vidar",
                "ID": 1,
                "IPAddress": "192.168.1.1",
                "InternalSSHPassword": "passw0rd",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 80,
                "Score": 1000,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 1,
                    "Logo": "https://vidar.club/logo.png",
                    "Name": "Vidar",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 1,
                "Visible": false
            },
            {
                "Challenge": {
                    "AutoRenewFlag": false,
                    "BaseScore": 1500,
                    "ID": 2,
                    "RenewFlagCommand": "",
                    "Title": "Web2"
                },
                "ChallengeID": 2,
                "Description": "Web2 for Vidar",
                "ID": 2,
                "IPAddress": "192.168.2.1",
                "InternalSSHPassword": "passw0rd",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 8080,
                "Score": 1500,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 1,
                    "Logo": "https://vidar.club/logo.png",
                    "Name": "Vidar",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 1,
                "Visible": false
            },
            {
                "Challenge": {
                    "AutoRenewFlag": false,
                    "BaseScore": 1500,
                    "ID": 2,
                    "RenewFlagCommand": "",
                    "Title": "Web2"
                },
                "ChallengeID": 2,
                "Description": "Web2 for E99p1ant",
                "ID": 3,
                "IPAddress": "192.168.2.2",
                "InternalSSHPassword": "passw0rd",
                "InternalSSHPort": 22,
                "InternalSSHUser": "root",
                "Port": 8080,
                "Score": 1500,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 2,
                    "Logo": "https://github.red/logo.png",
                    "Name": "E99p1ant",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 2,
                "Visible": false
            }
        ]
    },
    "error": 0
}
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/manager/gameBoxes", strings.NewReader(tc.body))
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
			ChallengeID:         1,
			TeamID:              1,
			IPAddress:           "192.168.1.1",
			Port:                80,
			Description:         "Web1 for Vidar",
			InternalSSHPort:     22,
			InternalSSHUser:     "root",
			InternalSSHPassword: "passw0rd",
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
			body: `
{
    "Description": "Web1 for Vidar-Team",
    "ID": 1,
    "IPAddress": "192.168.11.1",
	"Port": 80,
    "InternalSSHPassword": "s3cret",
    "InternalSSHPort": 2222,
    "InternalSSHUser": "root"
}
`,
			wantStatusCode: http.StatusOK,
			wantGameBoxList: `
{
    "data": {
        "Count": 1,
        "Data": [
            {
                "Challenge": {
                    "AutoRenewFlag": true,
                    "BaseScore": 1000,
                    "ID": 1,
                    "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                    "Title": "Web1"
                },
                "ChallengeID": 1,
                "Description": "Web1 for Vidar-Team",
                "ID": 1,
                "IPAddress": "192.168.11.1",
                "InternalSSHPassword": "s3cret",
                "InternalSSHPort": 2222,
                "InternalSSHUser": "root",
               	"Port": 80,
                "Score": 1000,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 1,
                    "Logo": "https://vidar.club/logo.png",
                    "Name": "Vidar",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 1,
                "Visible": false
            }
        ]
    },
    "error": 0
}
`,
		},
		{
			name: "game box not found",
			body: `
{
    "Description": "Web1 for Vidar-Team",
    "ID": 2,
    "IPAddress": "192.168.11.1",
	"Port": 80,
    "InternalSSHPassword": "s3cret",
    "InternalSSHPort": 2222,
    "InternalSSHUser": "root"
}
`,
			wantStatusCode: http.StatusNotFound,
			wantGameBoxList: `
{
    "data": {
        "Count": 1,
        "Data": [
            {
                "Challenge": {
                    "AutoRenewFlag": true,
                    "BaseScore": 1000,
                    "ID": 1,
                    "RenewFlagCommand": "echo {{FLAG}} \u003e /flag",
                    "Title": "Web1"
                },
                "ChallengeID": 1,
                "Description": "Web1 for Vidar-Team",
                "ID": 1,
                "IPAddress": "192.168.11.1",
                "InternalSSHPassword": "s3cret",
                "InternalSSHPort": 2222,
                "InternalSSHUser": "root",
               	"Port": 80,
                "Score": 1000,
                "IsDown": false,
				"IsCaptured": false,
                "Team": {
                    "ID": 1,
                    "Logo": "https://vidar.club/logo.png",
                    "Name": "Vidar",
                    "Rank": 1,
                    "Score": 0,
                    "Token": "mocked_randstr_hex"
                },
                "TeamID": 1,
                "Visible": false
            }
        ]
    },
    "error": 0
}
`,
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

	req, err := http.NewRequest(http.MethodPost, "/api/manager/gameBoxes", bytes.NewBuffer(bodyBytes))
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
