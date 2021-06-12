// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package cmd

import (
	"net/http"

	"github.com/Cardinal-Platform/binding"
	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/urfave/cli/v2"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/form"
	"github.com/vidar-team/Cardinal/internal/route"
	"github.com/vidar-team/Cardinal/internal/route/manager"
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

	if err = db.Init(); err != nil {
		log.Fatal("Failed to init database: %v", err)
	}

	f := flamego.Classic()

	f.Use(session.Sessioner(session.Options{
		ReadIDFunc:  func(r *http.Request) string { return r.Header.Get("Authorization") },
		WriteIDFunc: func(w http.ResponseWriter, r *http.Request, sid string, created bool) {},
	}))

	general := route.NewGeneralHandler()
	f.NotFound(general.NotFound)

	f.Group("/api", func() {
		f.Any("/", general.Hello)
		f.Get("/init", general.Init)
		f.Get("/time")
		f.Get("/uploads")
		f.Get("/asteroid")

		team := route.NewTeamHandler()
		f.Group("/team", func() {
			f.Post("/login", binding.Bind(form.TeamLogin{}), team.Login)
			f.Post("/logout", team.Logout)

			f.Group("", func() {
				f.Post("/submitFlag", team.SubmitFlag)
				f.Get("/info", team.Info)
				f.Get("/gameBoxes", team.GameBoxes)
				f.Get("/bulletins", team.Bulletins)
				f.Get("/liveLog")
			}, team.Authenticator)
		})

		manager := manager.NewHandler()
		f.Group("/manager", func() {
			// Panel
			f.Get("/panel")
			f.Get("/logs")
			f.Get("/rank")

			// Challenge
			f.Get("/challenges")
			f.Post("/challenge")
			f.Put("/challenge")
			f.Delete("/challenge")
			f.Post("/challenge/visible")

			// Team
			f.Get("/teams", manager.Teams)
			f.Post("/teams", manager.Teams)
			f.Put("/team", manager.UpdateTeam)
			f.Delete("/team", manager.DeleteTeam)
			f.Get("/team/resetPassword", manager.ResetTeamPassword)

			// Game Box
			f.Get("/gameBoxes")
			f.Post("/gameBoxes")
			f.Post("/gameBoxes/reset")
			f.Put("/gameBox")
			f.Post("/gameBox/sshTest")
			f.Post("/gameBox/refreshFlag")

			// Flag
			f.Get("/flags")
			f.Post("/flags")
			f.Get("/flags/export")

			// Bulletins
			f.Get("/bulletins")
			f.Post("/bulletin")
			f.Put("/bulletin")
			f.Delete("/bulletin")

			// Asteroid
			f.Group("/asteroid", func() {
				f.Get("/status")
				f.Post("/attack")
				f.Post("/rank")
				f.Post("/status")
				f.Post("/round")
				f.Post("/easterEgg")
				f.Post("/time")
				f.Post("/clear")
			})

			// Check
			f.Get("/checkDown")

		})
	})

	f.Use(context.Contexter())

	log.Info("Listen on http://0.0.0.0:%d", c.Int("port"))

	f.Run("0.0.0.0", c.Int("port"))
	return nil
}
