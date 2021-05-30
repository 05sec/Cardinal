// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package conf

import (
	"time"
)

// Build time and commit information.
// It should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

type Period struct {
	StartAt time.Time
	EndAt   time.Time
}

var (
	// App is the application settings.
	App struct {
		Version  string `ini:"-"` // Version should only be set by the main package.
		Language string
	}

	// Database is the database settings.
	Database struct {
		Type         string
		Host         string
		Name         string
		User         string
		Password     string
		SSLMode      string
		MaxOpenConns int
		MaxIdleConns int
	}

	// Game is the game settings.
	Game struct {
		Period    Period
		PauseTime []Period `ini:"PauseTime.Period,,,nonunique"`
		Duration  uint

		AttackScore    uint
		CheckDownScore uint
	}

	// Server is the web server settings.
	Server struct {
		HTTPAddr         string
		SeparateFrontend bool
	}
)
