package bootstrap

import (
	"github.com/vidar-team/Cardinal/internal/auth/manager"
	"github.com/vidar-team/Cardinal/internal/auth/team"
	"github.com/vidar-team/Cardinal/internal/bulletin"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/dynamic_config"
	"github.com/vidar-team/Cardinal/internal/game"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/misc/webhook"
)

func autoMigrate() {
	// Create tables.
	db.MySQL.AutoMigrate(
		&manager.Manager{},
		&game.Challenge{},
		&team.Token{},
		&team.Team{},
		&bulletin.Bulletin{},
		&bulletin.BulletinRead{}, // Not used

		&game.AttackAction{},
		&game.DownAction{},
		&game.Score{},
		&game.Flag{},
		&game.GameBox{},
		
		&logger.Log{},
		&webhook.WebHook{},

		&dynamic_config.DynamicConfig{},
	)
}
