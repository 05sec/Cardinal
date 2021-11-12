// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ FlagsStore = (*flags)(nil)

// Flags is the default instance of the FlagsStore.
var Flags FlagsStore

// FlagsStore is the persistent interface for flags.
type FlagsStore interface {
	// BatchCreate creates flags and persists to database.
	BatchCreate(ctx context.Context, opts CreateFlagOptions) error
	// Get returns the flags.
	Get(ctx context.Context, opts GetFlagOptions) ([]*Flag, int64, error)
	// Count counts the number of the flags with the given options.
	Count(ctx context.Context, opts CountFlagOptions) (int64, error)
	// Check checks the given flag.
	// It returns ErrFlagNotExists when not found.
	Check(ctx context.Context, flag string) (*Flag, error)
	// DeleteAll deletes all the flags.
	DeleteAll(ctx context.Context) error
}

// NewFlagsStore returns a FlagsStore instance with the given database connection.
func NewFlagsStore(db *gorm.DB) FlagsStore {
	return &flags{DB: db}
}

// Flag represents the flag which team submitted.
type Flag struct {
	gorm.Model

	TeamID      uint `gorm:"uniqueIndex:flag_unique_idx"`
	ChallengeID uint `gorm:"uniqueIndex:flag_unique_idx"`
	GameBoxID   uint `gorm:"uniqueIndex:flag_unique_idx"`
	Round       uint `gorm:"uniqueIndex:flag_unique_idx"`

	Value string
}

type flags struct {
	*gorm.DB
}

type FlagMetadata struct {
	GameBoxID uint
	Round     uint
	Value     string
}

type CreateFlagOptions struct {
	Flags []FlagMetadata
}

func (db *flags) BatchCreate(ctx context.Context, opts CreateFlagOptions) error {
	tx := db.WithContext(ctx).Begin()

	gameBoxIDs := map[uint]struct{}{}
	for _, flag := range opts.Flags {
		gameBoxIDs[flag.GameBoxID] = struct{}{}
	}

	var err error
	gameBoxesStore := NewGameBoxesStore(tx)
	gameBoxSets := make(map[uint]*GameBox, len(gameBoxIDs))
	for gameBoxID := range gameBoxIDs {
		gameBoxSets[gameBoxID], err = gameBoxesStore.GetByID(ctx, gameBoxID)
		if err != nil {
			tx.Rollback()
			if err == ErrGameBoxNotExists {
				return err
			}
			return errors.Wrap(err, "get game box")
		}
	}

	flags := make([]*Flag, 0, len(opts.Flags))
	for _, flag := range opts.Flags {
		flag := flag
		gameBox := gameBoxSets[flag.GameBoxID]
		flags = append(flags, &Flag{
			TeamID:      gameBox.TeamID,
			ChallengeID: gameBox.ChallengeID,
			GameBoxID:   gameBox.ID,
			Round:       flag.Round,
			Value:       flag.Value,
		})
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}, {Name: "challenge_id"}, {Name: "game_box_id"}, {Name: "round"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).CreateInBatches(flags, len(flags)).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "batch create flag")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

type GetFlagOptions struct {
	Page        int
	PageSize    int
	TeamID      uint
	ChallengeID uint
	GameBoxID   uint
	Round       uint
}

func (db *flags) Get(ctx context.Context, opts GetFlagOptions) ([]*Flag, int64, error) {
	var flags []*Flag
	var count int64

	q := db.WithContext(ctx).Model(&Flag{}).Where(&Flag{
		TeamID:      opts.TeamID,
		GameBoxID:   opts.GameBoxID,
		ChallengeID: opts.ChallengeID,
		Round:       opts.Round,
	})
	q.Count(&count)

	if opts.Page <= 0 {
		opts.Page = 1
	}

	if opts.PageSize != 0 {
		q = q.Offset((opts.Page - 1) * opts.PageSize).Limit(opts.PageSize)
	}

	return flags, count, q.Find(&flags).Error
}

type CountFlagOptions struct {
	TeamID      uint
	ChallengeID uint
	GameBoxID   uint
	Round       uint
}

func (db *flags) Count(ctx context.Context, opts CountFlagOptions) (int64, error) {
	var count int64
	q := db.WithContext(ctx).Model(&Flag{}).Where(&Flag{
		TeamID:      opts.TeamID,
		GameBoxID:   opts.GameBoxID,
		ChallengeID: opts.ChallengeID,
		Round:       opts.Round,
	})

	return count, q.Count(&count).Error
}

var ErrFlagNotExists = errors.New("flag does not find")

func (db *flags) Check(ctx context.Context, flagValue string) (*Flag, error) {
	var flag Flag
	err := db.WithContext(ctx).Model(&Flag{}).Where("value = ?", flagValue).First(&flag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrFlagNotExists
		}
		return nil, errors.Wrap(err, "get")
	}
	return &flag, nil
}

func (db *flags) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Flag{}).Error
}
