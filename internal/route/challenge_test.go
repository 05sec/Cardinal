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

func TestChallenge(t *testing.T) {
	t.Parallel()

	router, managerToken, cleanup := NewTestRoute(t)

	// Create one team.
	createTeam(t, managerToken, router, form.NewTeam{
		{
			Name: "Vidar",
			Logo: "https://vidar.club/logo.png",
		},
	})

	for _, tc := range []struct {
		name string
		test func(t *testing.T, router *flamego.Flame, managerToken string)
	}{
		{"List", testListChallenges},
		{"New", testNewChallenge},
		{"Update", testUpdateChallenge},
		{"Delete", testDeleteChallenge},
		{"SetVisible", testSetChallengeVisible},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("challenges")
				if err != nil {
					t.Fatal(err)
				}
			})

			tc.test(t, router, managerToken)
		})
	}
}

func testListChallenges(t *testing.T, router *flamego.Flame, managerToken string) {
	// Empty challenges.
	req, err := http.NewRequest(http.MethodGet, "/api/manager/challenges", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{"error":0,"data":[]}`
	assert.JSONEq(t, want, w.Body.String())

	// Create two challenges.
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "ShowHub",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: `echo "d3ctf{sh0whub_f1ag}" > /flag`,
	})
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "real_cloud",
		BaseScore:        1000,
		AutoRenewFlag:    false,
		RenewFlagCommand: "",
	})

	// Get the two challenges.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/challenges", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want = `
{
    "error": 0,
    "data": [
        {
            "Visible": false,
            "BaseScore": 1000,
            "AutoRenewFlag": true,
            "RenewFlagCommand": "echo \"d3ctf{sh0whub_f1ag}\" > /flag",
            "ID": 1,
            "Title": "ShowHub"
        },
        {
            "Title": "real_cloud",
            "Visible": false,
            "BaseScore": 1000,
            "AutoRenewFlag": false,
            "RenewFlagCommand": "",
            "ID": 2
        }
    ]
}
`
	assert.JSONPartialEq(t, want, w.Body.String())
}

func testNewChallenge(t *testing.T, router *flamego.Flame, managerToken string) {
	// Invalid JSON.
	req, err := http.NewRequest(http.MethodPost, "/api/manager/challenge", strings.NewReader(`{"Title: "ShowHub"`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40000,"msg":"Wrong Request Format!"}`, w.Body.String())

	// Missing title.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/challenge", strings.NewReader(`{"Score": 1000}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotNil(t, w.Body.String())

	// Field type error.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/challenge", strings.NewReader(`{"Title": "ShowHub", "BaseScore": "1234.56", "AutoRenewFlag": true, "RenewFlagCommand": "echo \"d3ctf{sh0whub_f1ag}\" > /flag"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40000,"msg":"Wrong Request Format!"}`, w.Body.String())

	// Normal JSON.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/challenge", strings.NewReader(`{"Title": "ShowHub", "BaseScore": 1234.56, "AutoRenewFlag": true, "RenewFlagCommand": "echo \"d3ctf{sh0whub_f1ag}\" > /flag"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"Create new challenge ShowHub succeed", "error":0}`, w.Body.String())

	// Challenge title repeated.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/challenge", strings.NewReader(`{"Title": "ShowHub", "BaseScore": 2345.67, "AutoRenewFlag": true}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"msg":"Duplicated Challenge Found!", "error":40000}`, w.Body.String())
}

func testUpdateChallenge(t *testing.T, router *flamego.Flame, managerToken string) {
	// Create two challenges.
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "ShowHub",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: `echo "d3ctf{sh0whub_f1ag}" > /flag`,
	})
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "real_cloud",
		BaseScore:        1000,
		AutoRenewFlag:    false,
		RenewFlagCommand: "",
	})

	// Invalid JSON.
	req, err := http.NewRequest(http.MethodPut, "/api/manager/challenge", strings.NewReader(`{"ID": 5, "Title": "Welcome"`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40000, "msg":"Wrong Request Format!"}`, w.Body.String())

	// Update not exist challenge.
	req, err = http.NewRequest(http.MethodPut, "/api/manager/challenge", strings.NewReader(`{"ID": 5, "Title": "ShowHub_Revenge", "BaseScore": 1500}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error": 40400, "msg":"Challenge Not Found!"}`, w.Body.String())

	// Update challenge.
	req, err = http.NewRequest(http.MethodPut, "/api/manager/challenge", strings.NewReader(`{"ID": 1, "Title": "ShowHub_Revenge", "BaseScore": 1500}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Check the challenges.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/challenges", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{
    "data": [
        {
            "ID": 1,
            "Title": "ShowHub_Revenge",
            "Visible": false,
            "BaseScore": 1500,
            "AutoRenewFlag": false,
            "RenewFlagCommand": ""
        },
        {
            "Title": "real_cloud",
            "Visible": false,
            "BaseScore": 1000,
            "AutoRenewFlag": false,
            "RenewFlagCommand": "",
            "ID": 2
        }
    ],
    "error": 0
}
`
	assert.JSONPartialEq(t, want, w.Body.String())
}

func testDeleteChallenge(t *testing.T, router *flamego.Flame, managerToken string) {
	// Create two challenges.
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "ShowHub",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: `echo "d3ctf{sh0whub_f1ag}" > /flag`,
	})
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "real_cloud",
		BaseScore:        1000,
		AutoRenewFlag:    false,
		RenewFlagCommand: "",
	})

	// Create a game box belongs to the first challenge.
	createGameBox(t, managerToken, router, form.NewGameBox{
		{
			ChallengeID:         1,
			TeamID:              1,
			IPAddress:           "192.168.1.1",
			Port:                22,
			Description:         "ShowHub for Vidar",
			InternalSSHPort:     22,
			InternalSSHUser:     "root",
			InternalSSHPassword: "passw0rd",
		},
	})

	// Delete not exist challenge.
	req, err := http.NewRequest(http.MethodDelete, "/api/manager/challenge?id=5", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":40400, "msg":"Challenge Not Found!"}`, w.Body.String())

	// Delete the first challenge.
	req, err = http.NewRequest(http.MethodDelete, "/api/manager/challenge?id=1", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// The game box of the challenge will also be deleted.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/gameBoxes", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":{"Count": 0, "Data": []}, "error":0}`, w.Body.String())

	// Only one challenge exists now.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/challenges", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{
    "data": [
        {
            "Title": "real_cloud",
            "Visible": false,
            "BaseScore": 1000,
            "AutoRenewFlag": false,
            "RenewFlagCommand": "",
            "ID": 2
        }
    ],
    "error": 0
}
`
	assert.JSONPartialEq(t, want, w.Body.String())
}

func testSetChallengeVisible(t *testing.T, router *flamego.Flame, managerToken string) {
	// Create two challenges.
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "ShowHub",
		BaseScore:        1000,
		AutoRenewFlag:    true,
		RenewFlagCommand: `echo "d3ctf{sh0whub_f1ag}" > /flag`,
	})
	createChallenge(t, managerToken, router, form.NewChallenge{
		Title:            "real_cloud",
		BaseScore:        1500,
		AutoRenewFlag:    false,
		RenewFlagCommand: "",
	})

	// Create the game boxes of the challenges.
	createGameBox(t, managerToken, router, form.NewGameBox{
		{
			ChallengeID:         1, // ShowHub
			TeamID:              1,
			IPAddress:           "192.168.1.1",
			Port:                80,
			Description:         "ShowHub",
			InternalSSHPort:     22,
			InternalSSHUser:     "root",
			InternalSSHPassword: "passw0rd",
		},
		{
			ChallengeID:         2, // real_cloud
			TeamID:              1,
			IPAddress:           "192.168.2.1",
			Port:                8080,
			Description:         "real_cloud",
			InternalSSHPort:     22,
			InternalSSHUser:     "root",
			InternalSSHPassword: "s3cret",
		},
	})

	// Invalid JSON.
	req, err := http.NewRequest(http.MethodPost, "/api/manager/challenge/visible", strings.NewReader(`{"ID": 1, "Visible": tr`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40000, "msg":"Wrong Request Format!"}`, w.Body.String())

	// Set the not exist challenge.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/challenge/visible", strings.NewReader(`{"ID": 5, "Visible": true}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":40400, "msg":"Challenge Not Found!"}`, w.Body.String())

	// Set the first challenge to public.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/challenge/visible", strings.NewReader(`{"ID": 1, "Visible": true}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Check the challenges.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/challenges", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{
    "data": [
        {
            "ID": 1,
            "Title": "ShowHub",
            "BaseScore": 1000,
			"Visible": true,
            "AutoRenewFlag": true,
            "RenewFlagCommand": "echo \"d3ctf{sh0whub_f1ag}\" > /flag"
        },
        {
            "ID": 2,
            "Title": "real_cloud",
            "Visible": false,
            "BaseScore": 1500,
            "AutoRenewFlag": false,
            "RenewFlagCommand": ""
        }
    ],
    "error": 0
}
`
	assert.JSONPartialEq(t, want, w.Body.String())

	// Set the second challenge to public.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/challenge/visible", strings.NewReader(`{"ID": 2, "Visible": true}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Check the challenges.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/challenges", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want = `{
    "data": [
        {
            "ID": 1,
            "Title": "ShowHub",
            "BaseScore": 1000,
			"Visible": true,
            "AutoRenewFlag": true,
            "RenewFlagCommand": "echo \"d3ctf{sh0whub_f1ag}\" > /flag"
        },
        {
            "ID": 2,
            "Title": "real_cloud",
            "Visible": true,
            "BaseScore": 1500,
            "AutoRenewFlag": false,
            "RenewFlagCommand": ""
        }
    ],
    "error": 0
}
`
	assert.JSONPartialEq(t, want, w.Body.String())
}

func createChallenge(t *testing.T, managerToken string, router *flamego.Flame, f form.NewChallenge) {
	bodyBytes, err := jsoniter.Marshal(f)
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, "/api/manager/challenge", bytes.NewBuffer(bodyBytes))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
