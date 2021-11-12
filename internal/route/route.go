// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/flamego/cors"
	"github.com/flamego/flamego"
	"github.com/flamego/session"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/form"
	"github.com/vidar-team/Cardinal/internal/i18n"
)

// NewRouter returns the router.
func NewRouter() *flamego.Flame {
	f := flamego.Classic()

	f.Use(
		session.Sessioner(session.Options{
			ReadIDFunc:  func(r *http.Request) string { return r.Header.Get("Authorization") },
			WriteIDFunc: func(w http.ResponseWriter, r *http.Request, sid string, created bool) {},
		}),
		context.Contexter(),
		i18n.I18n(),
		flamego.Static(flamego.StaticOptions{
			Directory: "uploads",
			Prefix:    "uploads",
		}),
	)

	f.Use(cors.CORS(
		cors.Options{
			Methods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		},
	))

	general := NewGeneralHandler()
	auth := NewAuthHandler()
	bulletin := NewBulletinHandler()
	challenge := NewChallengeHandler()
	flag := NewFlagHandler()
	gameBox := NewGameBoxHandler()
	team := NewTeamHandler()
	manager := NewManagerHandler()

	f.Group("/api", func() {
		f.Any("/", general.Hello)
		f.Get("/init", general.Init)
		f.Get("/time", general.Time)
		f.Get("/asteroid")

		f.Post("/submitFlag", form.Bind(form.SubmitFlag{}), auth.TeamTokenAuthenticator, team.SubmitFlag)

		f.Group("/team", func() {
			f.Post("/login", form.Bind(form.TeamLogin{}), auth.TeamLogin)
			f.Get("/logout", auth.TeamLogout)

			f.Group("", func() {
				f.Get("/info", team.Info)
				f.Get("/gameBoxes", func() {
					f.Get("/", team.GameBoxes)
					f.Get("/all")
				})
				f.Get("/bulletins", team.Bulletins)
				f.Get("/rank", team.Rank)
				f.Get("/liveLog")
			}, auth.TeamAuthenticator)
		})

		f.Group("/manager", func() {
			f.Post("/login", form.Bind(form.ManagerLogin{}), auth.ManagerLogin)
			f.Get("/logout", auth.ManagerLogout)

			f.Group("", func() {
				f.Get("/panel")
				f.Get("/logs")
				f.Get("/rank", manager.Rank)

				// Challenge
				f.Get("/challenges", challenge.List)
				f.Post("/challenge", form.Bind(form.NewChallenge{}), challenge.New)
				f.Put("/challenge", form.Bind(form.UpdateChallenge{}), challenge.Update)
				f.Delete("/challenge", challenge.Delete)
				f.Post("/challenge/visible", form.Bind(form.SetChallengeVisible{}), challenge.SetVisible)

				// Team
				f.Get("/teams", team.List)
				f.Post("/teams", form.Bind(form.NewTeam{}), team.New)
				f.Put("/team", form.Bind(form.UpdateTeam{}), team.Update)
				f.Delete("/team", team.Delete)
				f.Post("/team/resetPassword", team.ResetPassword)

				// Game Box
				f.Get("/gameBoxes", gameBox.List)
				f.Post("/gameBoxes/reset")
				f.Post("/gameBoxes", form.Bind(form.NewGameBox{}), gameBox.New)
				f.Put("/gameBox", form.Bind(form.UpdateGameBox{}), gameBox.Update)
				f.Delete("/gameBox", gameBox.Delete)
				f.Post("/gameBox/sshTest")
				f.Post("/gameBox/refreshFlag")

				// Flag
				f.Get("/flags", flag.Get)
				f.Post("/flags", flag.BatchCreate)

				// Bulletins
				f.Get("/bulletins", bulletin.List)
				f.Post("/bulletin", form.Bind(form.NewBulletin{}), bulletin.New)
				f.Put("/bulletin", form.Bind(form.UpdateBulletin{}), bulletin.Update)
				f.Delete("/bulletin", bulletin.Delete)

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

				// Account
				f.Group("/account", func() {
					f.Get("")
					f.Post("")
					f.Put("")
					f.Delete("")
				})

				// Check
				f.Get("/checkDown")
			}, auth.ManagerAuthenticator)
		})
	})

	f.NotFound(general.NotFound)

	return f
}
