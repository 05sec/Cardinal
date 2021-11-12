// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

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
	// BatchCreate creates game boxes in batch.
	// It returns the game boxes after they are created.
	BatchCreate(ctx context.Context, opts []CreateGameBoxOptions) ([]*GameBox, error)
	// Get returns the game boxes with the given options.
	Get(ctx context.Context, opts GetGameBoxesOption) ([]*GameBox, error)
	// GetByID returns the game box with given id.
	// It returns ErrGameBoxNotExists when not found.
	GetByID(ctx context.Context, id uint) (*GameBox, error)
	// Count returns the total count of game boxes.
	Count(ctx context.Context) (int64, error)
	// Update updates the game box with given id.
	Update(ctx context.Context, id uint, opts UpdateGameBoxOptions) error
	// SetScore updates the game box score with given id.
	SetScore(ctx context.Context, id uint, score float64) error
	// CountScore counts the game box total scores with the given options.
	CountScore(ctx context.Context, opts GameBoxCountScoreOptions) (float64, error)
	// SetVisible sets the game box visibility with given id.
	SetVisible(ctx context.Context, id uint, isVisible bool) error
	// SetDown activates the game box down status.
	SetDown(ctx context.Context, id uint) error
	// SetCaptured activates the game box captured status.
	SetCaptured(ctx context.Context, id uint) error
	// CleanStatus cleans the given game box's status.
	CleanStatus(ctx context.Context, id uint) error
	// CleanAllStatus sets all the game boxes' status to `GameBoxStatusUp`.
	CleanAllStatus(ctx context.Context) error
	// DeleteByIDs deletes the game box with given ids.
	DeleteByIDs(ctx context.Context, ids ...uint) error
	// DeleteAll deletes all the game boxes.
	DeleteAll(ctx context.Context) error
}

// NewGameBoxesStore returns a GameBoxesStore instance with the given database connection.
func NewGameBoxesStore(db *gorm.DB) GameBoxesStore {
	return &gameboxes{DB: db}
}

// GameBox represents the game box.
type GameBox struct {
	gorm.Model

	TeamID      uint
	Team        *Team `gorm:"-"`
	ChallengeID uint
	Challenge   *Challenge `gorm:"-"`

	IPAddress   string
	Port        uint
	Description string

	InternalSSHPort     uint
	InternalSSHUser     string
	InternalSSHPassword string

	Visible    bool
	Score      float64 // The score can be negative.
	IsDown     bool
	IsCaptured bool
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
	IPAddress   string
	Port        uint
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
	challenge, err := challengesStore.GetByID(ctx, opts.ChallengeID)
	if err != nil {
		if err == ErrChallengeNotExists {
			return 0, ErrChallengeNotExists
		}
		return 0, errors.Wrap(err, "get challenge")
	}

	var gameBox GameBox

	if err := db.WithContext(ctx).Model(&GameBox{}).Where("team_id = ? AND challenge_id = ?", opts.TeamID, opts.ChallengeID).First(&gameBox).Error; err == nil {
		return 0, ErrGameBoxAlreadyExists
	} else if err != gorm.ErrRecordNotFound {
		return 0, errors.Wrap(err, "get")
	}

	g := &GameBox{
		TeamID:              opts.TeamID,
		ChallengeID:         opts.ChallengeID,
		IPAddress:           opts.IPAddress,
		Port:                opts.Port,
		Description:         opts.Description,
		InternalSSHPort:     opts.InternalSSH.Port,
		InternalSSHUser:     opts.InternalSSH.User,
		InternalSSHPassword: opts.InternalSSH.Password,
		Score:               challenge.BaseScore,
	}

	if err := db.WithContext(ctx).Create(g).Error; err != nil {
		return 0, err
	}

	return g.ID, nil
}

func (db *gameboxes) BatchCreate(ctx context.Context, opts []CreateGameBoxOptions) ([]*GameBox, error) {
	tx := db.Begin()

	challengeIDSets := make(map[uint]struct{})
	for _, option := range opts {
		challengeIDSets[option.ChallengeID] = struct{}{}
	}
	challengeIDs := make([]uint, 0, len(challengeIDSets))
	for id := range challengeIDSets {
		challengeIDs = append(challengeIDs, id)
	}

	challengesStore := NewChallengesStore(db.DB)
	challenges, err := challengesStore.GetByIDs(ctx, challengeIDs...)
	if err != nil {
		return nil, errors.Wrap(err, "get challenges")
	}
	challengeSets := make(map[uint]*Challenge)
	for _, challenge := range challenges {
		challenge := challenge
		challengeSets[challenge.ID] = challenge
	}

	gameboxes := make([]*GameBox, 0, len(opts))
	for _, option := range opts {
		var gameBox GameBox
		if err := tx.WithContext(ctx).Model(&GameBox{}).Where("team_id = ? AND challenge_id = ?", option.TeamID, option.ChallengeID).First(&gameBox).Error; err == nil {
			tx.Rollback()
			return nil, ErrGameBoxAlreadyExists
		} else if err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return nil, errors.Wrap(err, "get")
		}

		g := &GameBox{
			TeamID:              option.TeamID,
			ChallengeID:         option.ChallengeID,
			IPAddress:           option.IPAddress,
			Port:                option.Port,
			Description:         option.Description,
			InternalSSHPort:     option.InternalSSH.Port,
			InternalSSHUser:     option.InternalSSH.User,
			InternalSSHPassword: option.InternalSSH.Password,
			Score:               challengeSets[option.ChallengeID].BaseScore,
		}
		if err := tx.WithContext(ctx).Create(g).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		gameboxes = append(gameboxes, g)
	}

	return gameboxes, tx.Commit().Error
}

func (db *gameboxes) loadAttributes(ctx context.Context, gameBoxes []*GameBox) ([]*GameBox, error) {
	teamIDs := map[uint]struct{}{}
	challengeIDs := map[uint]struct{}{}
	for _, gameBox := range gameBoxes {
		teamIDs[gameBox.TeamID] = struct{}{}
		challengeIDs[gameBox.ChallengeID] = struct{}{}
	}

	// Get game box's team.
	teamsStore := NewTeamsStore(db.DB)
	teamSets := map[uint]*Team{}
	for teamID := range teamIDs {
		var err error
		teamSets[teamID], err = teamsStore.GetByID(ctx, teamID)
		if err != nil {
			return nil, errors.Wrap(err, "get team")
		}
	}

	// Get game box's challenge.
	challengeStore := NewChallengesStore(db.DB)
	challengeSets := map[uint]*Challenge{}
	for challengeID := range challengeIDs {
		var err error
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

type GetGameBoxesOption struct {
	TeamID      uint
	ChallengeID uint
	Visible     bool // If Visible is `false`, it returns the visible and invisible game boxes.
	IsDown      bool
	IsCaptured  bool
}

func (db *gameboxes) Get(ctx context.Context, opts GetGameBoxesOption) ([]*GameBox, error) {
	var gameBoxes []*GameBox
	query := db.DB.WithContext(ctx).Model(&GameBox{}).Where(&GameBox{
		TeamID:      opts.TeamID,
		ChallengeID: opts.ChallengeID,
		Visible:     opts.Visible,
		IsDown:      opts.IsDown,
		IsCaptured:  opts.IsCaptured,
	})

	err := query.Order("id ASC").Find(&gameBoxes).Error
	if err != nil {
		return nil, err
	}

	return db.loadAttributes(ctx, gameBoxes)
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

	gameBoxes, err := db.loadAttributes(ctx, []*GameBox{&gameBox})
	if err != nil {
		return nil, errors.Wrap(err, "load attributes")
	}
	if len(gameBoxes) == 0 {
		return nil, errors.New("empty game boxes after loading attributes")
	}
	return gameBoxes[0], nil
}

func (db *gameboxes) Count(ctx context.Context) (int64, error) {
	var count int64
	return count, db.WithContext(ctx).Model(&GameBox{}).Count(&count).Error
}

type UpdateGameBoxOptions struct {
	IPAddress   string
	Port        uint
	Description string
	InternalSSH SSHConfig
}

func (db *gameboxes) Update(ctx context.Context, id uint, opts UpdateGameBoxOptions) error {
	var gameBox GameBox
	err := db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).First(&gameBox).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrGameBoxNotExists
		}
		return errors.Wrap(err, "get")
	}

	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).
		Updates(&GameBox{
			IPAddress:           opts.IPAddress,
			Port:                opts.Port,
			Description:         opts.Description,
			InternalSSHPort:     opts.InternalSSH.Port,
			InternalSSHUser:     opts.InternalSSH.User,
			InternalSSHPassword: opts.InternalSSH.Password,
		}).Error
}

func (db *gameboxes) SetScore(ctx context.Context, id uint, score float64) error {
	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).Update("score", score).Error
}

type GameBoxCountScoreOptions struct {
	TeamID      uint
	ChallengeID uint
	Visible     bool // If Visible is `false`, it returns the visible and invisible game boxes.
	IsDown      bool
	IsCaptured  bool
}

func (db *gameboxes) CountScore(ctx context.Context, opts GameBoxCountScoreOptions) (float64, error) {
	var sum struct {
		Score float64
	}

	return sum.Score, db.WithContext(ctx).Model(&GameBox{}).Select(`SUM(score) AS score`).Where(&GameBox{
		TeamID:      opts.TeamID,
		ChallengeID: opts.ChallengeID,
		Visible:     opts.Visible,
		IsDown:      opts.IsDown,
		IsCaptured:  opts.IsCaptured,
	}).Find(&sum).Error
}

func (db *gameboxes) SetVisible(ctx context.Context, id uint, isVisible bool) error {
	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).Update("visible", isVisible).Error
}

func (db *gameboxes) SetDown(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).Update("is_down", true).Error
}

func (db *gameboxes) SetCaptured(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).Update("is_captured", true).Error
}

func (db *gameboxes) CleanStatus(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Model(&GameBox{}).Where("id = ?", id).
		Update("is_down", false).
		Update("is_captured", false).
		Error
}

func (db *gameboxes) CleanAllStatus(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&GameBox{}).
		Update("is_down", false).
		Update("is_captured", false).
		Error
}

func (db *gameboxes) DeleteByIDs(ctx context.Context, id ...uint) error {
	return db.WithContext(ctx).Delete(&GameBox{}, "id IN (?)", id).Error
}

func (db *gameboxes) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&GameBox{}).Error
}
