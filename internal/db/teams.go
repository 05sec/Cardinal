// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

var _ TeamsStore = (*teams)(nil)

// Teams is the default instance of the TeamsStore.
var Teams TeamsStore

// TeamsStore is the persistent interface for teams.
type TeamsStore interface {
	// Authenticate validates name and password.
	// It returns ErrBadCredentials when validate failed.
	Authenticate(ctx context.Context, name, password string) (*Team, error)
	// Create creates a new team and persists to database.
	// It returns the team when it created.
	Create(ctx context.Context, opts CreateTeamOptions) (*Team, error)
	// BatchCreate creates teams in batch.
	// It returns the teams after they are created.
	BatchCreate(ctx context.Context, opts []CreateTeamOptions) ([]*Team, error)
	// Get returns the team list.
	Get(ctx context.Context, opts GetTeamsOptions) ([]*Team, error)
	// GetByID returns the team with given id.
	// It returns ErrTeamNotExists when not found.
	GetByID(ctx context.Context, id uint) (*Team, error)
	// GetByName returns the team with given name.
	// It returns ErrTeamNotExists when not found.
	GetByName(ctx context.Context, name string) (*Team, error)
	// ChangePassword changes the team's password with given id.
	ChangePassword(ctx context.Context, id uint, newPassword string) error
	// Update updates the team with given id.
	Update(ctx context.Context, id uint, opts UpdateTeamOptions) error
	// SetScore sets the team score with given id.
	SetScore(ctx context.Context, id uint, score float64) error
	// DeleteByID deletes the team with given id.
	DeleteByID(ctx context.Context, id uint) error
	// DeleteAll deletes all the teams.
	DeleteAll(ctx context.Context) error
}

// NewTeamsStore returns a TeamsStore instance with the given database connection.
func NewTeamsStore(db *gorm.DB) TeamsStore {
	return &teams{DB: db}
}

// Team represents the team.
type Team struct {
	gorm.Model

	Name     string
	Password string `json:"-"`
	Salt     string `json:"-"`
	Logo     string
	Score    float64
	Rank     uint
	Token    string
}

// EncodePassword encodes password to safe format.
func (t *Team) EncodePassword() {
	newPasswd := pbkdf2.Key([]byte(t.Password), []byte(t.Salt), 10000, 50, sha256.New)
	t.Password = fmt.Sprintf("%x", newPasswd)
}

// ValidatePassword checks if given password matches the one belongs to the team.
func (t *Team) ValidatePassword(password string) bool {
	newTeam := &Team{Password: password, Salt: t.Salt}
	newTeam.EncodePassword()
	return subtle.ConstantTimeCompare([]byte(t.Password), []byte(newTeam.Password)) == 1
}

// getTeamSalt returns a random team salt token.
func getTeamSalt() string {
	return randstr.String(10)
}

type teams struct {
	*gorm.DB
}

var ErrBadCredentials = errors.New("bad credentials")

func (db *teams) Authenticate(ctx context.Context, name, password string) (*Team, error) {
	var team Team
	if err := db.WithContext(ctx).Model(&Team{}).Where("name = ?", name).First(&team).Error; err != nil {
		return nil, ErrBadCredentials
	}

	if !team.ValidatePassword(password) {
		return nil, ErrBadCredentials
	}
	return &team, nil
}

type CreateTeamOptions struct {
	Name     string
	Password string
	Logo     string
}

var ErrTeamAlreadyExists = errors.New("team already exits")

func (db *teams) Create(ctx context.Context, opts CreateTeamOptions) (*Team, error) {
	var team Team
	if err := db.WithContext(ctx).Model(&Team{}).Where("name = ?", opts.Name).First(&team).Error; err == nil {
		return nil, ErrTeamAlreadyExists
	} else if err != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(err, "get")
	}

	t := &Team{
		Name:     opts.Name,
		Password: opts.Password,
		Salt:     getTeamSalt(),
		Logo:     opts.Logo,
		Token:    randstr.Hex(16), // Random token.
	}
	t.EncodePassword()

	return t, db.WithContext(ctx).Create(t).Error
}

func (db *teams) BatchCreate(ctx context.Context, opts []CreateTeamOptions) ([]*Team, error) {
	tx := db.Begin()

	teams := make([]*Team, 0, len(opts))
	for _, option := range opts {
		var team Team
		if err := tx.WithContext(ctx).Model(&Team{}).Where("name = ?", option.Name).First(&team).Error; err == nil {
			tx.Rollback()
			return nil, ErrTeamAlreadyExists
		} else if err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return nil, errors.Wrap(err, "get")
		}

		t := &Team{
			Name:     option.Name,
			Password: option.Password,
			Salt:     getTeamSalt(),
			Logo:     option.Logo,
			Token:    randstr.Hex(16), // Random token.
		}
		t.EncodePassword()
		if err := tx.WithContext(ctx).Create(t).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		teams = append(teams, t)
	}

	return teams, tx.Commit().Error
}

type GetTeamsOptions struct {
	OrderBy  string
	Page     int
	PageSize int
}

func (db *teams) Get(ctx context.Context, opts GetTeamsOptions) ([]*Team, error) {
	if opts.Page <= 0 {
		opts.Page = 1
	}

	if opts.OrderBy == "" {
		opts.OrderBy = "id ASC"
	}

	var teams []*Team
	return teams, db.WithContext(ctx).Model(&Team{}).
		Select([]string{"*", "RANK() OVER(ORDER BY score DESC) rank"}).
		Offset((opts.Page - 1) * opts.PageSize).
		Limit(opts.PageSize).
		Order(opts.OrderBy).Find(&teams).Error
}

var ErrTeamNotExists = errors.New("team dose not exist")

func (db *teams) GetByID(ctx context.Context, id uint) (*Team, error) {
	var team Team
	if err := db.WithContext(ctx).
		Table("(?) as teams",
			db.WithContext(ctx).Model(&Team{}).
				Select([]string{"*", "RANK() OVER(ORDER BY score DESC) rank"}),
		).
		Where("id = ?", id).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTeamNotExists
		}
		return nil, err
	}

	return &team, nil
}

func (db *teams) GetByName(ctx context.Context, name string) (*Team, error) {
	var team Team
	if err := db.WithContext(ctx).
		Table("(?) as teams",
			db.WithContext(ctx).Model(&Team{}).
				Select([]string{"*", "RANK() OVER(ORDER BY score DESC) rank"}),
		).
		Where("name = ?", name).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTeamNotExists
		}
		return nil, err
	}

	return &team, nil
}

func (db *teams) ChangePassword(ctx context.Context, id uint, newPassword string) error {
	var newTeam Team
	newTeam.Password = newPassword
	newTeam.EncodePassword()

	return db.WithContext(ctx).Model(&Team{}).Where("id = ?", id).Update("password", newTeam.Password).Error
}

type UpdateTeamOptions struct {
	Name  string
	Logo  string
	Token string
}

func (db *teams) Update(ctx context.Context, id uint, opts UpdateTeamOptions) error {
	return db.WithContext(ctx).Model(&Team{}).Where("id = ?", id).Updates(&Team{
		Name:  opts.Name,
		Logo:  opts.Logo,
		Token: opts.Token,
	}).Error
}

func (db *teams) SetScore(ctx context.Context, id uint, score float64) error {
	return db.WithContext(ctx).Model(&Team{}).Where("id = ?", id).Update("score", score).Error
}

func (db *teams) DeleteByID(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&Team{}, "id = ?", id).Error
}

func (db *teams) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Team{}).Error
}
