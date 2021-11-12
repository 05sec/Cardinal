// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/dbutil"
)

var AllTables = []interface{}{
	&Action{},
	&Bulletin{},
	&Challenge{},
	&Flag{},
	&GameBox{},
	&Log{},
	&Manager{},
	&Team{},
}

type DatabaseType string

const (
	DatabaseTypeMySQL    DatabaseType = "mysql"
	DatabaseTypePostgres DatabaseType = "postgres"
)

// Init initializes the database.
func Init() error {
	var dialector gorm.Dialector

	switch DatabaseType(conf.Database.Type) {
	case DatabaseTypeMySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4,utf8",
			conf.Database.User,
			conf.Database.Password,
			conf.Database.Host,
			conf.Database.Port,
			conf.Database.Name,
		)
		dialector = mysql.Open(dsn)

	case DatabaseTypePostgres:
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			conf.Database.User,
			conf.Database.Password,
			conf.Database.Host,
			conf.Database.Port,
			conf.Database.Name,
			conf.Database.SSLMode,
		)
		dialector = postgres.Open(dsn)

	default:
		log.Fatal("Unexpected database type: %q", conf.Database.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		NowFunc: func() time.Time {
			return dbutil.Now()
		},
	})
	if err != nil {
		return errors.Wrap(err, "open connection")
	}

	// Migrate databases.
	if db.AutoMigrate(AllTables...) != nil {
		return errors.Wrap(err, "auto migrate")
	}

	SetDatabaseStore(db)

	return nil
}

// SetDatabaseStore sets the database table store.
func SetDatabaseStore(db *gorm.DB) {
	Actions = NewActionsStore(db)
	Bulletins = NewBulletinsStore(db)
	Challenges = NewChallengesStore(db)
	Flags = NewFlagsStore(db)
	GameBoxes = NewGameBoxesStore(db)
	Ranks = NewRanksStore(db)
	Scores = NewScoresStore(db)
	Logs = NewLogsStore(db)
	Managers = NewManagersStore(db)
	Teams = NewTeamsStore(db)
}
