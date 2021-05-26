// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/flamego/flamego"
	"github.com/urfave/cli/v2"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/route/general"
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
	f := flamego.Classic()

	f.Group("/api", func() {
		f.Any("/", general.Hello)
	})

	f.NotFound(general.NotFound)

	f.Use(context.Contexter())

	log.Info("Listen on http://0.0.0.0:%d", c.Int("port"))

	f.Run("0.0.0.0", c.Int("port"))
	return nil
}
