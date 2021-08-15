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

	"github.com/flamego/flamego"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"

	"github.com/vidar-team/Cardinal/internal/form"
)

func TestChallenge(t *testing.T) {
	t.Parallel()

	router, managerToken, cleanup := NewTestRoute(t)

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
            "CreatedAt": "2020-01-09T10:06:40Z",
            "Title": "ShowHub"
        },
        {
            "CreatedAt": "2020-01-09T10:06:40Z",
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
	assert.JSONEq(t, want, w.Body.String())
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
	assert.JSONEq(t, `{"error":40001,"msg":"Challenge title is a required field"}`, w.Body.String())

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
            "CreatedAt": "2020-01-09T10:06:40Z",
            "Title": "ShowHub_Revenge",
            "Visible": false,
            "BaseScore": 1500,
            "AutoRenewFlag": false,
            "RenewFlagCommand": ""
        },
        {
            "CreatedAt": "2020-01-09T10:06:40Z",
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
	assert.JSONEq(t, want, w.Body.String())
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

	// Delete not exist challenge.
	req, err := http.NewRequest(http.MethodDelete, "/api/manager/challenge?id=5", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":40400, "msg":"Challenge Not Found!"}`, w.Body.String())

	// Delete the first bulletin.
	req, err = http.NewRequest(http.MethodDelete, "/api/manager/challenge?id=1", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Check the bulletins.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/challenges", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{
    "data": [
        {
            "CreatedAt": "2020-01-09T10:06:40Z",
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
	assert.JSONEq(t, want, w.Body.String())
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
		BaseScore:        1000,
		AutoRenewFlag:    false,
		RenewFlagCommand: "",
	})

	// TODO Create the game boxes of the challenges.
	createGameBox(t, managerToken, router, form.NewGameBox{
		ChallengeID: 0,
		TeamID:      0,
		Address:     "",
		Description: "",
		SSHPort:     0,
		SSHUser:     "",
		SSHPassword: "",
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
	assert.JSONEq(t, `{"error":40400, "msg":"Challenge does not exist."}`, w.Body.String())

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
            "RenewFlagCommand": "echo \"d3ctf{sh0whub_f1ag}\" > /flag",
            "CreatedAt": "2020-01-09T10:06:40Z"
        },
        {
            "ID": 2,
            "CreatedAt": "2020-01-09T10:06:40Z",
            "Title": "real_cloud",
            "Visible": false,
            "BaseScore": 1000,
            "AutoRenewFlag": false,
            "RenewFlagCommand": ""
        }
    ],
    "error": 0
}
`
	assert.JSONEq(t, want, w.Body.String())
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
