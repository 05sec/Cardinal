// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ BulletinsStore = (*bulletins)(nil)

// Bulletins is the default instance of the BulletinsStore.
var Bulletins BulletinsStore

// BulletinsStore is the persistent interface for bulletins.
type BulletinsStore interface {
	// Create creates a new bulletin and persists to database.
	// It returns the bulletin ID when bulletin created.
	Create(ctx context.Context, opts CreateBulletinOptions) (uint, error)
	// Get returns all the bulletins.
	Get(ctx context.Context) ([]*Bulletin, error)
	// GetByID returns the bulletin with given id.
	// It returns ErrBulletinNotExists when not found.
	GetByID(ctx context.Context, id uint) (*Bulletin, error)
	// Update updates the bulletin with given id.
	Update(ctx context.Context, id uint, opts UpdateBulletinOptions) error
	// DeleteByID deletes the bulletin with given id.
	DeleteByID(ctx context.Context, id uint) error
	// DeleteAll deletes all the bulletins.
	DeleteAll(ctx context.Context) error
}

// NewBulletinsStore returns a BulletinsStore instance with the given database connection.
func NewBulletinsStore(db *gorm.DB) BulletinsStore {
	return &bulletins{DB: db}
}

// Bulletin represents the bulletin which sent to teams.
type Bulletin struct {
	gorm.Model

	Title string
	Body  string
}

type bulletins struct {
	*gorm.DB
}

type CreateBulletinOptions struct {
	Title string
	Body  string
}

func (db *bulletins) Create(ctx context.Context, opts CreateBulletinOptions) (uint, error) {
	bulletin := &Bulletin{
		Title: opts.Title,
		Body:  opts.Body,
	}
	if err := db.WithContext(ctx).Create(bulletin).Error; err != nil {
		return 0, err
	}

	return bulletin.ID, nil
}

func (db *bulletins) Get(ctx context.Context) ([]*Bulletin, error) {
	var bulletins []*Bulletin
	return bulletins, db.DB.WithContext(ctx).Model(&Bulletin{}).Order("id ASC").Find(&bulletins).Error
}

var ErrBulletinNotExists = errors.New("bulletin does not exist")

func (db *bulletins) GetByID(ctx context.Context, id uint) (*Bulletin, error) {
	var bulletin Bulletin
	if err := db.WithContext(ctx).Model(&Bulletin{}).Where("id = ?", id).First(&bulletin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBulletinNotExists
		}
		return nil, errors.Wrap(err, "get")
	}
	return &bulletin, nil
}

type UpdateBulletinOptions struct {
	Title string
	Body  string
}

func (db *bulletins) Update(ctx context.Context, id uint, opts UpdateBulletinOptions) error {
	return db.WithContext(ctx).Model(&Bulletin{}).Where("id = ?", id).
		Select("Title", "Body").
		Updates(&Bulletin{
			Title: opts.Title,
			Body:  opts.Body,
		}).Error
}

func (db *bulletins) DeleteByID(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&Bulletin{}, "id = ?", id).Error
}

func (db *bulletins) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Bulletin{}).Error
}
