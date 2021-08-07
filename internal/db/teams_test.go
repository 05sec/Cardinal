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

func TestTeams(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := newTestDB(t)
	store := NewTeamsStore(db)

	for _, tc := range []struct {
		name string
		test func(t *testing.T, ctx context.Context, db *teams)
	}{
		{"Authenticate", testTeamsAuthenticate},
		{"Create", testTeamsCreate},
		{"Get", testTeamsGet},
		{"GetByID", testTeamsGetByID},
		{"GetByName", testTeamsGetByName},
		{"ChangePassword", testTeamsChangePassword},
		{"Update", testTeamsUpdate},
		{"SetScore", testTeamsSetScore},
		{"DeleteByID", testTeamsDeleteByID},
		{"DeleteAll", testTeamsDeleteAll},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				err := cleanup("teams")
				if err != nil {
					t.Fatal(err)
				}
			})
			tc.test(t, context.Background(), store.(*teams))
		})
	}
}

func testTeamsAuthenticate(t *testing.T, ctx context.Context, db *teams) {
	want, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
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
}

func testTeamsCreate(t *testing.T, ctx context.Context, db *teams) {
	got, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)
	assert.NotZero(t, got.Salt)
	assert.NotZero(t, got.Token)
	assert.NotZero(t, got.CreatedAt)
	assert.NotZero(t, got.UpdatedAt)

	got.Token = ""
	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &Team{
		Model: gorm.Model{
			ID: 1,
		},
		Name:     "Vidar",
		Password: "123456",
		Salt:     got.Salt,
		Logo:     "https://vidar.club/logo.png",
		Score:    0,
	}
	want.EncodePassword()

	assert.Equal(t, want, got)

	_, err = db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "asdfgh",
		Logo:     "https://vidar.club/logo_1.png",
	})
	assert.Equal(t, ErrTeamAlreadyExists, err)
}

func testTeamsGet(t *testing.T, ctx context.Context, db *teams) {
	team1, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)

	team2, err := db.Create(ctx, CreateTeamOptions{
		Name:     "E99p1ant",
		Password: "abcdef",
		Logo:     "https://github.red/",
	})
	assert.Nil(t, err)

	team3, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Cosmos",
		Password: "zxcvbn",
		Logo:     "https://cosmos.red/",
	})
	assert.Nil(t, err)

	got, err := db.Get(ctx, GetTeamsOptions{
		Page:     1,
		PageSize: 5,
	})

	for k := range got {
		got[k].CreatedAt = time.Time{}
		got[k].UpdatedAt = time.Time{}
	}

	assert.Nil(t, err)

	want := []*Team{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name:     "Vidar",
			Password: "123456",
			Salt:     team1.Salt,
			Logo:     "https://vidar.club/logo.png",
			Score:    0,
			Rank:     1,
			Token:    team1.Token,
		},
		{
			Model: gorm.Model{
				ID: 2,
			},
			Name:     "E99p1ant",
			Password: "abcdef",
			Salt:     team2.Salt,
			Logo:     "https://github.red/",
			Score:    0,
			Rank:     1,
			Token:    team2.Token,
		},
		{
			Model: gorm.Model{
				ID: 3,
			},
			Name:     "Cosmos",
			Password: "zxcvbn",
			Salt:     team3.Salt,
			Logo:     "https://cosmos.red/",
			Score:    0,
			Rank:     1,
			Token:    team3.Token,
		},
	}
	want[0].EncodePassword()
	want[1].EncodePassword()
	want[2].EncodePassword()
	assert.Equal(t, want, got)
}

func testTeamsGetByID(t *testing.T, ctx context.Context, db *teams) {
	want, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)
	want.Rank = 1

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, want, got)

	_, err = db.GetByID(ctx, 2)
	assert.Equal(t, ErrTeamNotExists, err)

	// Update score, test rank.
	_, err = db.Create(ctx, CreateTeamOptions{
		Name:     "E99p1ant",
		Password: "abcdef",
		Logo:     "https://github.red/",
	})
	assert.Nil(t, err)
	err = db.SetScore(ctx, 2, 1000)
	assert.Nil(t, err)

	got, err = db.GetByID(ctx, 2)
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want = &Team{
		Model: gorm.Model{
			ID: 2,
		},
		Name:     "E99p1ant",
		Password: "abcdef",
		Salt:     got.Salt,
		Logo:     "https://github.red/",
		Score:    1000,
		Rank:     1,
		Token:    got.Token,
	}
	want.EncodePassword()
	assert.Equal(t, want, got)

	got, err = db.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, uint(2), got.Rank)
}

func testTeamsGetByName(t *testing.T, ctx context.Context, db *teams) {
	want, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)
	want.Rank = 1

	got, err := db.GetByName(ctx, "Vidar")
	assert.Nil(t, err)
	assert.Equal(t, want, got)

	_, err = db.GetByName(ctx, "")
	assert.Equal(t, ErrTeamNotExists, err)

	// Update score, test rank.
	_, err = db.Create(ctx, CreateTeamOptions{
		Name:     "E99p1ant",
		Password: "abcdef",
		Logo:     "https://github.red/",
	})
	assert.Nil(t, err)
	err = db.SetScore(ctx, 2, 1000)
	assert.Nil(t, err)

	got, err = db.GetByName(ctx, "E99p1ant")
	assert.Nil(t, err)

	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want = &Team{
		Model: gorm.Model{
			ID: 2,
		},
		Name:     "E99p1ant",
		Password: "abcdef",
		Salt:     got.Salt,
		Logo:     "https://github.red/",
		Score:    1000,
		Rank:     1,
		Token:    got.Token,
	}
	want.EncodePassword()
	assert.Equal(t, want, got)

	got, err = db.GetByName(ctx, "Vidar")
	assert.Nil(t, err)
	assert.Equal(t, uint(2), got.Rank)
}

func testTeamsChangePassword(t *testing.T, ctx context.Context, db *teams) {
	team, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)

	oldPassword := team.Password

	err = db.ChangePassword(ctx, 1, "newp@ssword")
	assert.Nil(t, err)

	team, err = db.GetByID(ctx, 1)
	assert.Nil(t, err)
	newPassword := team.Password

	assert.NotEqual(t, oldPassword, newPassword)

	err = db.ChangePassword(ctx, 2, "user_not_found")
	assert.Nil(t, err)
}

func testTeamsUpdate(t *testing.T, ctx context.Context, db *teams) {
	team, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)

	err = db.Update(ctx, 1, UpdateTeamOptions{
		Name:  "Vidar-Team",
		Logo:  "https://vidar.club/new_logo.png",
		Token: "new_t0ken",
	})
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)
	got.CreatedAt = time.Time{}
	got.UpdatedAt = time.Time{}

	want := &Team{
		Model: gorm.Model{
			ID: 1,
		},
		Name:     "Vidar-Team",
		Password: "123456",
		Salt:     team.Salt,
		Logo:     "https://vidar.club/new_logo.png",
		Score:    0,
		Rank:     1,
		Token:    "new_t0ken",
	}
	want.EncodePassword()
	assert.Equal(t, want, got)
}

func testTeamsSetScore(t *testing.T, ctx context.Context, db *teams) {
	_, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)

	err = db.SetScore(ctx, 1, 1500)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, got.Score, float64(1500))
}

func testTeamsDeleteByID(t *testing.T, ctx context.Context, db *teams) {
	_, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)

	err = db.DeleteByID(ctx, 1)
	assert.Nil(t, err)

	got, err := db.GetByID(ctx, 1)
	assert.Equal(t, ErrTeamNotExists, err)
	want := (*Team)(nil)
	assert.Equal(t, want, got)
}

func testTeamsDeleteAll(t *testing.T, ctx context.Context, db *teams) {
	_, err := db.Create(ctx, CreateTeamOptions{
		Name:     "Vidar",
		Password: "123456",
		Logo:     "https://vidar.club/logo.png",
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateTeamOptions{
		Name:     "E99p1ant",
		Password: "abcdef",
		Logo:     "https://github.red/",
	})
	assert.Nil(t, err)

	_, err = db.Create(ctx, CreateTeamOptions{
		Name:     "Cosmos",
		Password: "zxcvbn",
		Logo:     "https://cosmos.red/",
	})
	assert.Nil(t, err)

	err = db.DeleteAll(ctx)
	assert.Nil(t, err)

	got, err := db.Get(ctx, GetTeamsOptions{})
	assert.Nil(t, err)
	want := []*Team{}
	assert.Equal(t, want, got)
}
