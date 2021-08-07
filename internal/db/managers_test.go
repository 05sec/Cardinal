// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestManagers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)
	store := NewManagersStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *managers)
	}{
		{"Authenticate", testManagersAuthenticate},
		{"Create", testManagersCreate},
		{"Get", testManagersGet},
		{"GetByID", testManagersGetByID},
		{"ChangePassword", testManagersChangePassword},
		{"Update", testManagersUpdate},
		{"DeleteByID", testManagersDeleteByID},
		{"DeleteAll", testManagersDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("managers")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), store.(*managers))
		})
	}
}

func testManagersAuthenticate(t *testing.T, ctx context.Context, db *managers) {
	want, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Vidar",
		Password:       "123456",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	// Correct credential.
	got, err := db.Authenticate(ctx, "Vidar", "123456")
	assert.Nil(t, err)
	assert.Equal(t, want, got)

	// Incorrect password.
	_, err = db.Authenticate(ctx, "Vidar", "abcdef")
	assert.Equal(t, ErrBadCredentials, err)

	// Empty credential.
	_, err = db.Authenticate(ctx, "", "")
	assert.Equal(t, ErrBadCredentials, err)

	_, err = db.Create(ctx, CreateManagerOptions{
		Name:           "Checker",
		Password:       "qwerty",
		IsCheckAccount: true,
	})
	assert.Nil(t, err)

	_, err = db.Authenticate(ctx, "Checker", "qwerty")
	assert.Equal(t, ErrBadCredentials, err)
}

func testManagersCreate(t *testing.T, ctx context.Context, db *managers) {
	got, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Vidar",
		Password:       "123456",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)
	assert.NotZero(t, got.Salt)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &Manager{
		Model: gorm.Model{
			ID: 1,
		},
		Name:           "Vidar",
		Password:       "123456",
		Salt:           got.Salt,
		IsCheckAccount: false,
	}
	want.EncodePassword()

	assert.Equal(t, want, got)

	_, err = db.Create(ctx, CreateManagerOptions{
		Name:           "Vidar",
		Password:       "abcdef",
		IsCheckAccount: true,
	})
	assert.Equal(t, ErrManagerAlreadyExists, err)
}

func testManagersGet(t *testing.T, ctx context.Context, db *managers) {
	manager1, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Vidar",
		Password:       "123456",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	manager2, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Checker",
		Password:       "abcdef",
		IsCheckAccount: true,
	})
	assert.Nil(t, err)

	manager3, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Cosmos",
		Password:       "zxcvbn",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	got, err := db.Get(ctx)

	for k := range got {
		got[k].CreatedAt = time.Time{}
		got[k].UpdatedAt = time.Time{}
	}

	assert.Nil(t, err)

	want := []*Manager{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name:           "Vidar",
			Password:       "123456",
			Salt:           manager1.Salt,
			IsCheckAccount: false,
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Name:           "Checker",
			Password:       "abcdef",
			Salt:           manager2.Salt,
			IsCheckAccount: true,
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			Name:           "Cosmos",
			Password:       "zxcvbn",
			Salt:           manager3.Salt,
			IsCheckAccount: false,
		},
	}
	want[0].EncodePassword()
	want[1].EncodePassword()
	want[2].EncodePassword()
	assert.Equal(t, want, got)
}

func testManagersGetByID(t *testing.T, ctx context.Context, db *managers) {
	_, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Vidar",
		Password:       "123456",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &Manager{
		Model: gorm.Model{
			ID: 1,
		},
		Name:           "Vidar",
		Password:       "123456",
		Salt:           got.Salt,
		IsCheckAccount: false,
	}
	want.EncodePassword()
	assert.Equal(t, want, got)

	// Get not exist manager.
	got, err = db.GetByID(ctx, 2)
	assert.Equal(t, ErrManagerNotExists, err)
	want = (*Manager)(nil)
	assert.Equal(t, want, got)
}

func testManagersChangePassword(t *testing.T, ctx context.Context, db *managers) {
	manager, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Vidar",
		Password:       "123456",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	oldPassword := manager.Password

	err = db.ChangePassword(ctx, 1, "newp@ssword")
	assert.Nil(t, err)

	manager, err = db.GetByID(ctx, 1)
	assert.Nil(t, err)
	newPassword := manager.Password

	assert.NotEqual(t, oldPassword, newPassword)

	err = db.ChangePassword(ctx, 2, "user_not_found")
	assert.Nil(t, err)
}

func testManagersUpdate(t *testing.T, ctx context.Context, db *managers) {
	manager, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Checker",
		Password:       "abcdef",
		IsCheckAccount: true,
	})
	assert.Nil(t, err)

	err = db.Update(ctx, 1, UpdateManagerOptions{
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)
	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &Manager{
		Model: gorm.Model{
			ID: 1,
		},
		Name:           "Checker",
		Password:       "abcdef",
		Salt:           manager.Salt,
		IsCheckAccount: false,
	}
	want.EncodePassword()
	assert.Equal(t, want, got)
}

func testManagersDeleteByID(t *testing.T, ctx context.Context, db *managers) {
	_, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Vidar",
		Password:       "123456",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	err = db.DeleteByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Equal(t, ErrManagerNotExists, err)
	want := (*Manager)(nil)
	assert.Equal(t, want, got)
}

func testManagersDeleteAll(t *testing.T, ctx context.Context, db *managers) {
	_, err := db.Create(ctx, CreateManagerOptions{
		Name:           "Vidar",
		Password:       "123456",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateManagerOptions{
		Name:           "Checker",
		Password:       "abcdef",
		IsCheckAccount: true,
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateManagerOptions{
		Name:           "Cosmos",
		Password:       "zxcvbn",
		IsCheckAccount: false,
	})
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	got, err := db.Get(ctx)
	assert.Nil(t, err)
	want := []*Manager{}
	assert.Equal(t, want, got)
}
