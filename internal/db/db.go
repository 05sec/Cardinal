// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package db

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vidar-team/Cardinal/internal/dbold"
	"github.com/vidar-team/Cardinal/internal/dbutil"
)

var ErrBadCharset = errors.New("bad charset")

var allTables = []interface{}{
	&Bulletin{},
	&Challenge{},
	&Manager{},
	&Team{},
}

// Init initializes the database.
func Init(username, password, host, port, name, sslMode string) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", username, password, host, port, name, sslMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return dbutil.Now()
		},
	})
	if err != nil {
		return errors.Wrap(err, "open connection")
	}

	// Migrate databases.
	if db.AutoMigrate(allTables...) != nil {
		return errors.Wrap(err, "auto migrate")
	}

	// Test database charset, we should support Chinese input.
	if dbold.MySQL.Exec("SELECT * FROM `logs` WHERE `Content` = '中文测试';").Error != nil {
		return ErrBadCharset
	}

	Bulletins = NewBulletinsStore(db)
	Challenges = NewChallengesStore(db)
	Teams = NewTeamsStore(db)

	return nil
}
