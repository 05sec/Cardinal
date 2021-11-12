// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/urfave/cli/v2"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/clock"
	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/route"
	"github.com/vidar-team/Cardinal/internal/store"
)

var Web = &cli.Command{
	Name:  "web",
	Usage: "Start web server",
	Description: `Cardinal web server is the only thing you need to run,
and it takes care of all the other things for you`,
	Action: runWeb,
	Flags: []cli.Flag{
		intFlag("port, p", 19999, "Temporary port number to prevent conflict"),
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func runWeb(c *cli.Context) error {
	err := conf.Init(c.String("config"))
	if err != nil {
		log.Fatal("Failed to load config: %v", err)
	}
	log.Trace(locales.T("config.load_success"))

	if err = db.Init(); err != nil {
		log.Fatal("Failed to init database: %v", err)
	}

	// TODO Install

	store.Init()

	if err := clock.Init(); err != nil {
		log.Fatal("Failed to init clock: %v", err)
	}
	clock.Start()

	f := route.NewRouter()
	log.Info("Listen on http://0.0.0.0:%d", c.Int("port"))

	f.Run("0.0.0.0", c.Int("port"))
	return nil
}
