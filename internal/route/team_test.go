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

func TestTeam(t *testing.T) {
	router, managerToken, cleanup := NewTestRoute(t)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, router *flamego.Flame, managerToken string)
	}{
		{"List", testListTeams},
		{"New", testNewTeam},
		{"Update", testUpdateTeam},
		{"Delete", testDeleteTeam},
		{"ResetPassword", testResetPasswordTeam},
		//{"SubmitFlag"},
		//{"Info"},
		//{"GameBoxes"},
		//{"Bulletins"},
		//{"Rank"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("teams")
				if err != nil {
					t.Fatal(err)
				}
			})

			tc.test(t, router, managerToken)
		})
	}
}

func testListTeams(t *testing.T, router *flamego.Flame, managerToken string) {
	// Empty teams.
	req, err := http.NewRequest(http.MethodGet, "/api/manager/teams", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{"error":0,"data":[]}`
	assert.JSONEq(t, want, w.Body.String())

	// Create two teams.
	createTeam(t, managerToken, router, form.NewTeam{
		{
			Name: "Vidar",
			Logo: "https://vidar.club/logo.png",
		},
		{
			Name: "E99p1ant",
			Logo: "https://github.red/",
		},
	})

	// Get the two teams.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/teams", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want = `{
    "data": [
        {
            "ID": 1,
            "Logo": "https://vidar.club/logo.png",
            "Name": "Vidar",
            "Rank": 1,
            "Score": 0,
            "Token": "mocked_randstr_hex"
        },
        {
            "ID": 2,
            "Logo": "https://github.red/",
            "Name": "E99p1ant",
            "Rank": 1,
            "Score": 0,
            "Token": "mocked_randstr_hex"
        }
    ],
    "error": 0
}`
	assert.JSONPartialEq(t, want, w.Body.String())
}

func testNewTeam(t *testing.T, router *flamego.Flame, managerToken string) {
	req, err := http.NewRequest(http.MethodPost, "/api/manager/teams", strings.NewReader(`[{"Name": "Vidar", "Logo": "https://vidar.club/logo.png"}]`))
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{
    "error": 0,
    "data": [
        {
            "Name": "Vidar",
            "Logo": "https://vidar.club/logo.png"
        }
    ]
}`
	assert.JSONPartialEq(t, want, w.Body.String())

	// Create the same team repeatedly.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/teams", strings.NewReader(`[{"Name": "Vidar", "Logo": "https://vidar.club/logo.png"}]`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40000,"msg":"Duplicated Team Found!"}`, w.Body.String())

	// Create teams in batch.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/teams", strings.NewReader(`[{"Name": "E99p1ant", "Logo": "https://github.red/logo.png"}, {"Name": "Cosmos", "Logo": "https://cosmos.red/logo.png"}]`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want = `{
    "error": 0,
    "data": [
        {
            "Name": "E99p1ant",
            "Logo": "https://github.red/logo.png"
        },
		{
            "Name": "Cosmos",
            "Logo": "https://cosmos.red/logo.png"
        }
    ]
}`
	assert.JSONPartialEq(t, want, w.Body.String())
}

func testUpdateTeam(t *testing.T, router *flamego.Flame, managerToken string) {
	// Create two teams.
	createTeam(t, managerToken, router, form.NewTeam{
		{Name: "Vidar", Logo: "https://vidar.club/logo.png"},
		{Name: "E99p1ant", Logo: "https://github.red/logo.png"},
	})

	// Update team E99p1ant.
	req, err := http.NewRequest(http.MethodPut, "/api/manager/team", strings.NewReader(`{"ID": 2, "Name": "John", "Logo": "https://github.red/John.png"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Update not exist team.
	req, err = http.NewRequest(http.MethodPut, "/api/manager/team", strings.NewReader(`{"ID": 3, "Name": "None", "Logo": "https://github.red/John.png"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Update `Logo` only.
	req, err = http.NewRequest(http.MethodPut, "/api/manager/team", strings.NewReader(`{"ID": 1, "Name": "Vidar", "Logo": "https://vidar.club/logo_new.png"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Set exist team name.
	req, err = http.NewRequest(http.MethodPut, "/api/manager/team", strings.NewReader(`{"ID": 1, "Name": "John", "Logo": "https://vidar.club/logo_new.png"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40000, "msg":"Team name \"John\" repeat."}`, w.Body.String())
}

func testDeleteTeam(t *testing.T, router *flamego.Flame, managerToken string) {
	// Create two teams.
	createTeam(t, managerToken, router, form.NewTeam{
		{Name: "Vidar", Logo: "https://vidar.club/logo.png"},
		{Name: "E99p1ant", Logo: "https://github.red/logo.png"},
	})

	// Delete team E99p1ant.
	req, err := http.NewRequest(http.MethodDelete, "/api/manager/team?id=2", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Make sure the team has been deleted.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/teams", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{
    "data": [
        {
            "ID": 1,
            "Logo": "https://vidar.club/logo.png",
            "Name": "Vidar",
            "Rank": 1,
            "Score": 0,
            "Token": "mocked_randstr_hex"
        }
    ],
    "error": 0
}
`
	assert.JSONPartialEq(t, want, w.Body.String())

	// Delete not exists team.
	req, err = http.NewRequest(http.MethodDelete, "/api/manager/team?id=3", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":40400, "msg":"Team Not Found!"}`, w.Body.String())
}

func testResetPasswordTeam(t *testing.T, router *flamego.Flame, managerToken string) {
	// Create two teams.
	createTeam(t, managerToken, router, form.NewTeam{
		{Name: "Vidar", Logo: "https://vidar.club/logo.png"},
		{Name: "E99p1ant", Logo: "https://github.red/logo.png"},
	})

	// Reset team E99p1ant password.
	req, err := http.NewRequest(http.MethodPost, "/api/manager/team/resetPassword?id=2", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONPartialEq(t, `{"error":0}`, w.Body.String())

	// Reset not exist team password.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/team/resetPassword?id=3", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONPartialEq(t, `{"error":40400}`, w.Body.String())
}

func createTeam(t *testing.T, managerToken string, router *flamego.Flame, f form.NewTeam) {
	bodyBytes, err := jsoniter.Marshal(f)
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, "/api/manager/teams", bytes.NewBuffer(bodyBytes))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotZero(t, w.Body.String())
}
