// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ LogsStore = (*logs)(nil)

// Logs is the default instance of the LogsStore.
var Logs LogsStore

// LogsStore is the persistent interface for logs.
type LogsStore interface {
	// Create creates a new log and persists to database.
	Create(ctx context.Context, opts CreateLogOptions) error
	// Get returns all the logs.
	Get(ctx context.Context) ([]*Log, error)
	// DeleteAll deletes all the logs.
	DeleteAll(ctx context.Context) error
}

// NewLogsStore returns a LogsStore instance with the given database connection.
func NewLogsStore(db *gorm.DB) LogsStore {
	return &logs{DB: db}
}

type LogLevel int

const (
	LogLevelNormal LogLevel = iota
	LogLevelWarning
	LogLevelImportant
)

type LogType string

const (
	LogTypeHealthCheck    LogType = "health_check"
	LogTypeManagerOperate LogType = "manager_operate"
	LogTypeSSH            LogType = "ssh"
	LogTypeSystem         LogType = "system"
)

// Log represents the log.
type Log struct {
	gorm.Model

	Level LogLevel
	Type  LogType
	Body  string
}

type logs struct {
	*gorm.DB
}

type CreateLogOptions struct {
	Level LogLevel
	Type  LogType
	Body  string
}

var ErrBadLogLevel = errors.New("bad log level")
var ErrBadLogType = errors.New("bad log type")

func (db *logs) Create(ctx context.Context, opts CreateLogOptions) error {
	switch opts.Level {
	case LogLevelNormal, LogLevelWarning, LogLevelImportant:
	default:
		return ErrBadLogLevel
	}

	switch opts.Type {
	case LogTypeHealthCheck, LogTypeManagerOperate, LogTypeSSH, LogTypeSystem:
	default:
		return ErrBadLogType
	}

	return db.WithContext(ctx).Create(&Log{
		Level: opts.Level,
		Type:  opts.Type,
		Body:  opts.Body,
	}).Error
}

func (db *logs) Get(ctx context.Context) ([]*Log, error) {
	var logs []*Log
	return logs, db.WithContext(ctx).Model(&Log{}).Order("id DESC").Find(&logs).Error
}

func (db *logs) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Log{}).Error
}
