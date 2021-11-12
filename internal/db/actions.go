// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ ActionsStore = (*actions)(nil)

// Actions is the default instance of the ActionsStore.
var Actions ActionsStore

// ActionsStore is the persistent interface for actions.
type ActionsStore interface {
	// Create creates a new action and persists to database, it returns the Action if succeeded.
	Create(ctx context.Context, opts CreateActionOptions) (*Action, error)
	// Get returns the actions according to the given options.
	Get(ctx context.Context, opts GetActionOptions) ([]*Action, error)
	// SetScore updates the action's score.
	SetScore(ctx context.Context, opts SetActionScoreOptions) error
	// CountScore counts score with the given options.
	CountScore(ctx context.Context, opts CountActionScoreOptions) (float64, error)
	// GetEmptyScore returns the empty score actions in the given round.
	GetEmptyScore(ctx context.Context, round uint, actionType ActionType) ([]*Action, error)
	// Delete deletes the actions with the given options.
	Delete(ctx context.Context, opts DeleteActionOptions) error
	// DeleteAll deletes all the actions.
	DeleteAll(ctx context.Context) error
}

// NewActionsStore returns a ActionsStore instance with the given database connection.
func NewActionsStore(db *gorm.DB) ActionsStore {
	return &actions{DB: db}
}

type ActionType uint

const (
	ActionTypeBeenAttack ActionType = iota
	ActionTypeCheckDown
	ActionTypeAttack
	ActionTypeServiceOnline
)

// Action represents the action such as check down or being attacked.
type Action struct {
	gorm.Model

	Type           ActionType `gorm:"uniqueIndex:action_unique_idx"`
	TeamID         uint       `gorm:"uniqueIndex:action_unique_idx"`
	ChallengeID    uint       `gorm:"uniqueIndex:action_unique_idx"`
	GameBoxID      uint       `gorm:"uniqueIndex:action_unique_idx"`
	AttackerTeamID uint       `gorm:"uniqueIndex:action_unique_idx"`
	Round          uint       `gorm:"uniqueIndex:action_unique_idx"`

	Score float64
}

type actions struct {
	*gorm.DB
}

type CreateActionOptions struct {
	Type           ActionType
	GameBoxID      uint
	AttackerTeamID uint
	Round          uint
}

var ErrDuplicateAction = errors.New("duplicate action")

func (db *actions) Create(ctx context.Context, opts CreateActionOptions) (*Action, error) {
	if opts.Type == ActionTypeCheckDown || opts.Type == ActionTypeAttack {
		opts.AttackerTeamID = 0
	}

	gameBoxStore := NewGameBoxesStore(db.DB)
	gameBox, err := gameBoxStore.GetByID(ctx, opts.GameBoxID)
	if err != nil {
		return nil, err
	}

	tx := db.WithContext(ctx).Begin()
	var action Action
	err = tx.Model(&Action{}).Where(&Action{
		Type:           opts.Type,
		TeamID:         gameBox.TeamID,
		ChallengeID:    gameBox.ChallengeID,
		GameBoxID:      gameBox.ID,
		AttackerTeamID: opts.AttackerTeamID,
		Round:          opts.Round,
	}).First(&action).Error
	if err == nil {
		tx.Rollback()
		return nil, ErrDuplicateAction
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, errors.Wrap(err, "get action")
	}

	action = Action{
		Type:           opts.Type,
		TeamID:         gameBox.TeamID,
		ChallengeID:    gameBox.ChallengeID,
		GameBoxID:      gameBox.ID,
		AttackerTeamID: opts.AttackerTeamID,
		Round:          opts.Round,
	}
	err = tx.Create(&action).Error
	if err != nil {
		tx.Rollback()

		// NOTE: How to check if error type is DUPLICATE KEY in GORM.
		// https://github.com/go-gorm/gorm/issues/4037

		// Postgres
		if pgError, ok := err.(*pgconn.PgError); ok && errors.Is(err, pgError) && pgError.Code == "23505" {
			return nil, ErrDuplicateAction
		}
		// MySQL
		var mysqlErr mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrDuplicateAction
		}
		return nil, err
	}

	return &action, tx.Commit().Error
}

type GetActionOptions struct {
	ActionID       uint
	Type           ActionType
	TeamID         uint
	ChallengeID    uint
	GameBoxID      uint
	AttackerTeamID uint
	Round          uint
}

func (db *actions) Get(ctx context.Context, opts GetActionOptions) ([]*Action, error) {
	var actions []*Action
	return actions, db.WithContext(ctx).Model(&Action{}).Where(&Action{
		Model:          gorm.Model{ID: opts.ActionID},
		Type:           opts.Type,
		TeamID:         opts.TeamID,
		ChallengeID:    opts.ChallengeID,
		GameBoxID:      opts.GameBoxID,
		AttackerTeamID: opts.AttackerTeamID,
		Round:          opts.Round,
	}).Find(&actions).Error
}

type SetActionScoreOptions struct {
	ActionID  uint
	Round     uint
	GameBoxID uint
	Score     float64

	Replace bool
}

var ErrActionScoreInvalid = errors.New("invalid score, please check the action type and the sign of the score")

func (db *actions) SetScore(ctx context.Context, opts SetActionScoreOptions) error {
	actions, err := db.Get(ctx, GetActionOptions{
		ActionID:  opts.ActionID,
		GameBoxID: opts.GameBoxID,
		Round:     opts.Round,
	})
	if err != nil {
		return err
	}

	if len(actions) == 0 {
		return nil
	}

	action := actions[0]

	// Check the action score sign, the BeenAttack and CheckDown score must be negative,
	// the Attack and ServiceOnline score must be positive.
	if action.Type == ActionTypeBeenAttack || action.Type == ActionTypeCheckDown {
		if opts.Score > 0 {
			return ErrActionScoreInvalid
		}
	} else if action.Type == ActionTypeAttack || action.Type == ActionTypeServiceOnline {
		if opts.Score < 0 {
			return ErrActionScoreInvalid
		}
	}

	// If the score is not zero, we prefer not updating it, only if `replace` is true.
	if action.Score == 0 || opts.Replace {
		return db.WithContext(ctx).Model(&Action{}).Where("id = ?", action.ID).Update("score", opts.Score).Error
	}
	return nil
}

type CountActionScoreOptions struct {
	Type           ActionType
	TeamID         uint
	ChallengeID    uint
	GameBoxID      uint
	AttackerTeamID uint
	Round          uint
}

func (db *actions) CountScore(ctx context.Context, opts CountActionScoreOptions) (float64, error) {
	var sum struct {
		Score float64
	}

	return sum.Score, db.WithContext(ctx).Model(&Action{}).Select(`SUM(score) AS score`).Where(&Action{
		Type:           opts.Type,
		TeamID:         opts.TeamID,
		ChallengeID:    opts.ChallengeID,
		GameBoxID:      opts.GameBoxID,
		AttackerTeamID: opts.AttackerTeamID,
		Round:          opts.Round,
	}).Find(&sum).Error
}

func (db *actions) GetEmptyScore(ctx context.Context, round uint, actionType ActionType) ([]*Action, error) {
	var actions []*Action
	return actions, db.WithContext(ctx).Model(&Action{}).Where("round = ? AND type = ? AND score = 0", round, actionType).Find(&actions).Error
}

type DeleteActionOptions struct {
	ActionID       uint
	Type           ActionType
	TeamID         uint
	ChallengeID    uint
	GameBoxID      uint
	AttackerTeamID uint
	Round          uint
}

func (db *actions) Delete(ctx context.Context, opts DeleteActionOptions) error {
	return db.WithContext(ctx).Where(&Action{
		Model: gorm.Model{
			ID: opts.ActionID,
		},
		Type:           opts.Type,
		TeamID:         opts.TeamID,
		ChallengeID:    opts.ChallengeID,
		GameBoxID:      opts.GameBoxID,
		AttackerTeamID: opts.AttackerTeamID,
		Round:          opts.Round,
	}).Delete(&Action{}).Error
}

func (db *actions) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Action{}).Error
}
