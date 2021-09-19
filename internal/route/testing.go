// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"bou.ke/monkey"
	"github.com/flamego/flamego"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/thanhpk/randstr"

	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/dbutil"
	"github.com/vidar-team/Cardinal/internal/form"
	"github.com/vidar-team/Cardinal/internal/utils"
)

const (
	TestRouteAdminName     = "cardinal_admin"
	TestRouteAdminPassword = "supersecurepassword"
)

// NewTestRoute returns the route used for test.
func NewTestRoute(t *testing.T) (*flamego.Flame, string, func(tables ...string) error) {
	f := NewRouter()
	testDB, dbCleanup := dbutil.NewTestDB(t, db.AllTables...)
	// Set the global database store to test database.
	db.SetDatabaseStore(testDB)

	// Mock utils.GenerateToken()
	tokenPatch := monkey.Patch(utils.GenerateToken, func() string { return "mocked_token" })
	t.Cleanup(func() {
		tokenPatch.Unpatch()
	})

	// Mock github.com/thanhpk/randstr package.
	randstrPatch := monkey.Patch(randstr.Hex, func(int) string { return "mocked_randstr_hex" })
	t.Cleanup(func() {
		randstrPatch.Unpatch()
	})

	// Create manager account for testing.
	ctx := context.Background()
	_, err := db.Managers.Create(ctx, db.CreateManagerOptions{
		Name:           TestRouteAdminName,
		Password:       TestRouteAdminPassword,
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	// Login as manager to get the manager token.
	loginBody, err := jsoniter.Marshal(form.ManagerLogin{
		Name:     TestRouteAdminName,
		Password: TestRouteAdminPassword,
	})
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/manager/login", bytes.NewBuffer(loginBody))
	assert.Nil(t, err)
	f.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	respBody := struct {
		Data string `json:"data"`
	}{}
	err = jsoniter.NewDecoder(w.Body).Decode(&respBody)
	assert.Nil(t, err)
	managerToken := respBody.Data

	return f, managerToken, func(tables ...string) error {
		if t.Failed() {
			return nil
		}

		// Reset database table.
		if err := dbCleanup(tables...); err != nil {
			return err
		}
		return nil
	}
}
