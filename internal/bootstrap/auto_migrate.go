package bootstrap

import (
	"github.com/vidar-team/Cardinal/internal/db"
)

func autoMigrate() {
	// Create tables.
	db.MySQL.AutoMigrate(
		&db.Manager{},
		&db.Challenge{},
		&db.Token{},
		&db.Team{},
		&db.Bulletin{},
		&db.BulletinRead{}, // Not used

		&db.AttackAction{},
		&db.DownAction{},
		&db.Score{},
		&db.Flag{},
		&db.GameBox{},

		&db.Log{},
		&db.WebHook{},

		&db.DynamicConfig{},
	)
}
