// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package conf

import (
	"os"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/thanhpk/randstr"
)

func TestInit() error {
	config, err := toml.Load(
		`
[App]
Language = "zh-CN"
HTTPAddr = ":19999"
SeparateFrontend = false
EnableSentry = false

[Database]
Type = "mysql"
SSLMode = "disable"
MaxOpenConns = 50
MaxIdleConns = 50

[Game]
Duration = 300
AttackScore = 10
CheckDownScore = 10
`)
	if err != nil {
		return errors.Wrap(err, "load test config")
	}

	if err := parse(config); err != nil {
		return errors.Wrap(err, "parse config")
	}

	App.SecuritySalt = randstr.String(64)

	// Connect to the test environment database.
	Database.Host = os.ExpandEnv("$DBHOST:$DBPORT")
	Database.User = os.Getenv("DBUSER")
	Database.Password = os.Getenv("DBPASSWORD")
	Database.Name = os.Getenv("DBNAME")

	Game.StartAt = toml.LocalDateTimeOf(time.Now())
	Game.EndAt = toml.LocalDateTimeOf(time.Now().Add(12 * time.Hour))

	return nil
}
