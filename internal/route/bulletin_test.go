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

func TestBulletin(t *testing.T) {
	t.Parallel()

	router, managerToken, cleanup := NewTestRoute(t)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, router *flamego.Flame, managerToken string)
	}{
		{"List", testListBulletins},
		{"New", testNewBulletins},
		{"Update", testUpdateBulletins},
		{"Delete", testDeleteBulletins},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("bulletins")
				if err != nil {
					t.Fatal(err)
				}
			})

			tc.test(t, router, managerToken)
		})
	}
}

func testListBulletins(t *testing.T, router *flamego.Flame, managerToken string) {
	// Empty bulletins.
	req, err := http.NewRequest(http.MethodGet, "/api/manager/bulletins", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{"error":0,"data":[]}`
	assert.JSONEq(t, want, w.Body.String())

	// Create two bulletins.
	createBulletin(t, managerToken, router, "Welcome", "Welcome to D^3CTF!")
	createBulletin(t, managerToken, router, "Hint for Web1", "/web.zip")

	// Get the two bulletins.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/bulletins", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want = `{
    "data": [
        {
            "UpdatedAt": "2020-01-09T10:06:40Z",
            "DeletedAt": null,
            "Title": "Welcome",
            "Body": "Welcome to D^3CTF!",
            "ID": 1,
            "CreatedAt": "2020-01-09T10:06:40Z"
        },
        {
            "Body": "/web.zip",
            "ID": 2,
            "CreatedAt": "2020-01-09T10:06:40Z",
            "UpdatedAt": "2020-01-09T10:06:40Z",
            "DeletedAt": null,
            "Title": "Hint for Web1"
        }
    ],
    "error": 0
}
`
	assert.JSONEq(t, want, w.Body.String())
}

func testNewBulletins(t *testing.T, router *flamego.Flame, managerToken string) {
	// Invalid JSON.
	req, err := http.NewRequest(http.MethodPost, "/api/manager/bulletin", strings.NewReader(`{"Title": "No body"`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40000, "msg":"Wrong Request Format!"}`, w.Body.String())

	// Missing bulletin body.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/bulletin", strings.NewReader(`{"Title": "No body"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40001, "msg":"Bulletin Body is a required field"}`, w.Body.String())

	// Normal JSON.
	req, err = http.NewRequest(http.MethodPost, "/api/manager/bulletin", strings.NewReader(`{"Title": "Welcome", "Body": "Welcome to D^3CTF!"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"error": 0, "data": ""}`, w.Body.String())
}

func testUpdateBulletins(t *testing.T, router *flamego.Flame, managerToken string) {
	// Create two bulletins.
	createBulletin(t, managerToken, router, "Welcome", "Welcome to D^3CTF!")
	createBulletin(t, managerToken, router, "Hint for Web1", "/web.zip")

	// Invalid JSON.
	req, err := http.NewRequest(http.MethodPut, "/api/manager/bulletin", strings.NewReader(`{"ID": 5, "Title": "Welcome", "Body": "Wel`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":40000, "msg":"Wrong Request Format!"}`, w.Body.String())

	// Update not exist bulletin.
	req, err = http.NewRequest(http.MethodPut, "/api/manager/bulletin", strings.NewReader(`{"ID": 5, "Title": "Welcome", "Body": "Welcome to D^3CTF!"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error": 40400, "msg":"Bulletin Not Found!"}`, w.Body.String())

	// Update bulletin.
	req, err = http.NewRequest(http.MethodPut, "/api/manager/bulletin", strings.NewReader(`{"ID": 1, "Title": "Welcome!!", "Body": "Welcome to HCTF!"}`))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Check the bulletins.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/bulletins", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{
    "data": [
        {
            "UpdatedAt": "2020-01-09T10:06:40Z",
            "DeletedAt": null,
            "Title": "Welcome!!",
            "Body": "Welcome to HCTF!",
            "ID": 1,
            "CreatedAt": "2020-01-09T10:06:40Z"
        },
        {
            "Body": "/web.zip",
            "ID": 2,
            "CreatedAt": "2020-01-09T10:06:40Z",
            "UpdatedAt": "2020-01-09T10:06:40Z",
            "DeletedAt": null,
            "Title": "Hint for Web1"
        }
    ],
    "error": 0
}
`
	assert.JSONEq(t, want, w.Body.String())
}

func testDeleteBulletins(t *testing.T, router *flamego.Flame, managerToken string) {
	// Create two bulletins.
	createBulletin(t, managerToken, router, "Welcome", "Welcome to D^3CTF!")
	createBulletin(t, managerToken, router, "Hint for Web1", "/web.zip")

	// Delete not exist bulletin.
	req, err := http.NewRequest(http.MethodDelete, "/api/manager/bulletin?id=5", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":40400, "msg":"Bulletin Not Found!"}`, w.Body.String())

	// Delete the first bulletin.
	req, err = http.NewRequest(http.MethodDelete, "/api/manager/bulletin?id=1", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"data":"", "error":0}`, w.Body.String())

	// Check the bulletins.
	req, err = http.NewRequest(http.MethodGet, "/api/manager/bulletins", nil)
	assert.Nil(t, err)

	req.Header.Set("Authorization", managerToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	want := `{
    "data": [
        {
            "Body": "/web.zip",
            "ID": 2,
            "CreatedAt": "2020-01-09T10:06:40Z",
            "UpdatedAt": "2020-01-09T10:06:40Z",
            "DeletedAt": null,
            "Title": "Hint for Web1"
        }
    ],
    "error": 0
}
`
	assert.JSONEq(t, want, w.Body.String())
}

func createBulletin(t *testing.T, managerToken string, router *flamego.Flame, title, body string) {
	f := form.NewBulletin{
		Title: title,
		Body:  body,
	}
	bodyBytes, err := jsoniter.Marshal(f)
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, "/api/manager/bulletin", bytes.NewBuffer(bodyBytes))
	assert.Nil(t, err)
	req.Header.Set("Authorization", managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"error": 0, "data": ""}`, w.Body.String())
}
