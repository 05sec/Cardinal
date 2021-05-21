// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
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

var _ ManagersStore = (*managers)(nil)

// Managers is the default instance of the ManagersStore.
var Managers ManagersStore

// ManagersStore is the persistent interface for managers.
type ManagersStore interface {
	// Authenticate validates name and password.
	// It returns ErrBadCredentials when validate failed.
	// The check account can't log in.
	Authenticate(ctx context.Context, name, password string) (*Manager, error)
	// Create creates a new manager and persists to database.
	// It returns the manager when it created.
	Create(ctx context.Context, opts CreateManagerOptions) (*Manager, error)
	// Get returns all the managers.
	Get(ctx context.Context) ([]*Manager, error)
	// GetByID returns the manager with given id.
	// It returns ErrManagerNotExists when not found.
	GetByID(ctx context.Context, id uint) (*Manager, error)
	// ChangePassword changes the manager's password with given id.
	ChangePassword(ctx context.Context, id uint, newPassword string) error
	// Update updates the manager with given id.
	Update(ctx context.Context, id uint, opts UpdateManagerOptions) error
	// DeleteByID deletes the manager with given id.
	DeleteByID(ctx context.Context, id uint) error
	// DeleteAll deletes all the managers.
	DeleteAll(ctx context.Context) error
}

// NewManagersStore returns a ManagersStore instance with the given database connection.
func NewManagersStore(db *gorm.DB) ManagersStore {
	return &managers{DB: db}
}

// Manager represents the manager.
type Manager struct {
	gorm.Model

	Name           string
	Salt           string
	Password       string
	IsCheckAccount bool
}

// EncodePassword encodes password to safe format.
func (m *Manager) EncodePassword() {
	newPasswd := pbkdf2.Key([]byte(m.Password), []byte(m.Salt), 10000, 50, sha256.New)
	m.Password = fmt.Sprintf("%x", newPasswd)
}

// ValidatePassword checks if given password matches the one belongs to the manager.
func (m *Manager) ValidatePassword(password string) bool {
	newManager := &Manager{Password: password, Salt: m.Salt}
	newManager.EncodePassword()
	return subtle.ConstantTimeCompare([]byte(m.Password), []byte(newManager.Password)) == 1
}

// getManagerSalt returns a random manager salt token.
func getManagerSalt() string {
	return randstr.String(10)
}

type managers struct {
	*gorm.DB
}

func (db *managers) Authenticate(ctx context.Context, name, password string) (*Manager, error) {
	var manager Manager
	if err := db.WithContext(ctx).Model(&Manager{}).Where("name = ?", name).First(&manager).Error; err != nil {
		return nil, ErrBadCredentials
	}

	// Check account can't log in.
	if manager.IsCheckAccount || !manager.ValidatePassword(password) {
		return nil, ErrBadCredentials
	}
	return &manager, nil
}

type CreateManagerOptions struct {
	Name           string
	Password       string
	IsCheckAccount bool
}

var ErrManagerAlreadyExists = errors.New("manager already exits")

func (db *managers) Create(ctx context.Context, opts CreateManagerOptions) (*Manager, error) {
	var manager Manager
	if err := db.WithContext(ctx).Model(&Manager{}).Where("name = ?", opts.Name).First(&manager).Error; err == nil {
		return nil, ErrManagerAlreadyExists
	} else if err != gorm.ErrRecordNotFound {
		return nil, errors.Wrap(err, "get")
	}

	m := &Manager{
		Name:           opts.Name,
		Password:       opts.Password,
		Salt:           getManagerSalt(),
		IsCheckAccount: opts.IsCheckAccount,
	}
	m.EncodePassword()

	return m, db.WithContext(ctx).Create(m).Error
}

func (db *managers) Get(ctx context.Context) ([]*Manager, error) {
	var managers []*Manager
	return managers, db.WithContext(ctx).Model(&Manager{}).Order("id ASC").Find(&managers).Error
}

var ErrManagerNotExists = errors.New("manager dose not exist")

func (db *managers) GetByID(ctx context.Context, id uint) (*Manager, error) {
	var manager Manager
	if err := db.WithContext(ctx).Model(&Manager{}).Where("id = ?", id).First(&manager).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrManagerNotExists
		}
		return nil, err
	}

	return &manager, nil
}

func (db *managers) ChangePassword(ctx context.Context, id uint, newPassword string) error {
	var newManager Manager
	newManager.Password = newPassword
	newManager.EncodePassword()

	return db.WithContext(ctx).Model(&Manager{}).Where("id = ?", id).Update("password", newManager.Password).Error
}

type UpdateManagerOptions struct {
	IsCheckAccount bool
}

func (db *managers) Update(ctx context.Context, id uint, opts UpdateManagerOptions) error {
	return db.WithContext(ctx).Model(&Manager{}).Where("id = ?", id).Update("is_check_account", opts.IsCheckAccount).Error
}

func (db *managers) DeleteByID(ctx context.Context, id uint) error {
	return db.WithContext(ctx).Delete(&Manager{}, "id = ?", id).Error
}

func (db *managers) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Manager{}).Error
}
