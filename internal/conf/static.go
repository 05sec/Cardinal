// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package conf

import (
	"github.com/pelletier/go-toml"
)

// Build time and commit information.
// It should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

type Period struct {
	StartAt toml.LocalDateTime
	EndAt   toml.LocalDateTime
}

var (
	// App is the application settings.
	App struct {
		Version          string `toml:"-"` // Version should only be set by the main package.
		Language         string
		HTTPAddr         string
		SeparateFrontend bool
		EnableSentry     bool
		SecuritySalt     string
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
		StartAt   toml.LocalDateTime
		EndAt     toml.LocalDateTime
		PauseTime []Period
		Duration  uint

		AttackScore    uint
		CheckDownScore uint
	}
)
