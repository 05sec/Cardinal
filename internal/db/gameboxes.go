// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ GameBoxesStore = (*gameboxes)(nil)

// GameBoxes is the default instance of the GameBoxesStore.
var GameBoxes GameBoxesStore

// GameBoxesStore is the persistent interface for game boxes.
type GameBoxesStore interface {
	// Create creates a new game box and persists to database.
	// It returns the game box ID when game box created.
	Create(ctx context.Context, opts CreateGameBoxOptions) (uint, error)
	// Get returns all the game boxes.
	Get(ctx context.Context) ([]*GameBox, error)
	// GetByID returns the game box with given id.
	// It returns ErrGameBoxNotExists when not found.
	GetByID(ctx context.Context, id uint) (*GameBox, error)
	// Update updates the game box with given id.
	Update(ctx context.Context, id uint, opts UpdateGameBoxOptions) error
	// SetScore updates the game box score with given id.
	SetScore(ctx context.Context, id uint, score float64) error
	// SetVisible sets the game box visibility with given id.
	SetVisible(ctx context.Context, id uint, isVisible bool) error
	// SetStatus sets the game box status with given id.
	SetStatus(ctx context.Context, id uint, status GameBoxStatus) error
	// DeleteByID deletes the game box with given id.
	DeleteByID(ctx context.Context, id uint) error
	// DeleteAll deletes all the game boxes.
	DeleteAll(ctx context.Context) error
}

// NewGameBoxesStore returns a GameBoxesStore instance with the given database connection.
func NewGameBoxesStore(db *gorm.DB) GameBoxesStore {
	return &gameboxes{DB: db}
}

type GameBoxStatus string

const (
	GameBoxStatusUp       = "up"
	GameBoxStatusDown     = "down"
	GameBoxStatusCaptured = "captured"
)

// GameBox represents the game box.
type GameBox struct {
	gorm.Model

	TeamID      uint
	Team        *Team `db:"-"`
	ChallengeID uint
	Challenge   *Challenge `db:"-"`

	Address     string
	Description string

	InternalSSHPort     string
	InternalSSHUser     string
	InternalSSHPassword string

	Visible bool
	Score   float64 // The score can be negative.
	Status  GameBoxStatus
}

type gameboxes struct {
	*gorm.DB
}

type SSHConfig struct {
	Port     uint
	User     string
	Password string
}

type CreateGameBoxOptions struct {
	TeamID      uint
	ChallengeID uint
	Address     string
	Description string
	InternalSSH SSHConfig
}

var ErrGameBoxAlreadyExists = errors.New("game box already exits")

func (db *gameboxes) Create(ctx context.Context, opts CreateGameBoxOptions) (uint, error) {
	teamsStore := NewTeamsStore(db.DB)
	if _, err := teamsStore.GetByID(ctx, opts.TeamID); err != nil {
		if err == ErrTeamNotExists {
			return 0, ErrTeamNotExists
		}
		return 0, errors.Wrap(err, "get team")
	}

	challengesStore := NewChallengesStore(db.DB)
	if _, err := challengesStore.GetByID(ctx, opts.ChallengeID); err != nil {
		if err == ErrChallengeNotExists {
			return 0, ErrChallengeNotExists
		}
		return 0, errors.Wrap(err, "get challenge")
	}

	var gameBox GameBox

	if err := db.WithContext(ctx).Where("team_id = ? AND challenge_id = ?", opts.TeamID, opts.ChallengeID).First(&gameBox).Error; err == nil {
		return 0, ErrGameBoxAlreadyExists
	} else if err != gorm.ErrRecordNotFound {
		return 0, errors.Wrap(err, "get")
	}

	g := &GameBox{
		TeamID:              opts.TeamID,
		ChallengeID:         opts.ChallengeID,
		Address:             opts.Address,
		Description:         opts.Description,
		InternalSSHPort:     strconv.Itoa(int(opts.InternalSSH.Port)),
		InternalSSHUser:     opts.InternalSSH.User,
		InternalSSHPassword: opts.InternalSSH.Password,
		Visible:             false,
		Score:               0,
		Status:              GameBoxStatusUp,
	}

	if err := db.WithContext(ctx).Create(g).Error; err != nil {
		return 0, err
	}

	return g.ID, nil
}

func (db *gameboxes) Get(ctx context.Context) ([]*GameBox, error) {
	var gameBoxes []*GameBox
	err := db.DB.WithContext(ctx).Model(&GameBox{}).Order("id ASC").Find(&gameBoxes).Error
	if err != nil {
		return nil, err
	}

	teamIDs := map[uint]struct{}{}
	challengeIDs := map[uint]struct{}{}
	for _, gameBox := range gameBoxes {
		teamIDs[gameBox.TeamID] = struct{}{}
		challengeIDs[gameBox.ChallengeID] = struct{}{}
	}

	teamsStore := NewTeamsStore(db.DB)
	teamSets := map[uint]*Team{}
	for teamID := range teamIDs {
		teamSets[teamID], err = teamsStore.GetByID(ctx, teamID)
		if err != nil {
			return nil, errors.Wrap(err, "get team")
		}
	}

	challengeStore := NewChallengesStore(db.DB)
	challengeSets := map[uint]*Challenge{}
	for challengeID := range challengeIDs {
		challengeSets[challengeID], err = challengeStore.GetByID(ctx, challengeID)
		if err != nil {
			return nil, errors.Wrap(err, "get challenge")
		}
	}

	for _, gameBox := range gameBoxes {
		gameBox.Team = teamSets[gameBox.TeamID]
		gameBox.Challenge = challengeSets[gameBox.ChallengeID]
	}

	return gameBoxes, nil
}

var ErrGameBoxNotExists = errors.New("game box does not exist")

func (db *gameboxes) GetByID(ctx context.Context, id uint) (*GameBox, error) {
	var gameBox GameBox
	err := db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).First(&gameBox).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrGameBoxNotExists
		}
		return nil, errors.Wrap(err, "get")
	}

	teamsStore := NewTeamsStore(db.DB)
	gameBox.Team, err = teamsStore.GetByID(ctx, gameBox.TeamID)
	if err != nil {
		return nil, errors.Wrap(err, "get team")
	}

	challengeStore := NewChallengesStore(db.DB)
	gameBox.Challenge, err = challengeStore.GetByID(ctx, gameBox.ChallengeID)
	if err != nil {
		return nil, errors.Wrap(err, "get challenge")
	}

	return &gameBox, nil
}

type UpdateGameBoxOptions struct {
	Address     string
	Description string
	InternalSSH SSHConfig
}

func (db *gameboxes) Update(ctx context.Context, id uint, opts UpdateGameBoxOptions) error {
	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).
		Updates(&GameBox{
			Address:             opts.Address,
			Description:         opts.Description,
			InternalSSHPort:     strconv.Itoa(int(opts.InternalSSH.Port)),
			InternalSSHUser:     opts.InternalSSH.User,
			InternalSSHPassword: opts.InternalSSH.Password,
		}).Error
}

func (db *gameboxes) SetScore(ctx context.Context, id uint, score float64) error {
	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).Update("score", score).Error
}

func (db *gameboxes) SetVisible(ctx context.Context, id uint, isVisible bool) error {
	return db.WithContext(ctx).Model(&GameBox{}).Where("visible = ?", id).Update("visible", isVisible).Error
}

var ErrBadGameBoxsStatus = errors.New("bad game box status")

func (db *gameboxes) SetStatus(ctx context.Context, id uint, status GameBoxStatus) error {
	switch status {
	case GameBoxStatusUp, GameBoxStatusDown, GameBoxStatusCaptured:
	default:
		return ErrBadGameBoxsStatus
	}

	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).Update("status", status).Error
}

func (db *gameboxes) DeleteByID(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&GameBox{}, "id = ?", id).Error
}

func (db *gameboxes) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&GameBox{}).Error
}
