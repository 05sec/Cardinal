// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"time"

	"github.com/flamego/session"
	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/form"
)

// TeamHandler is the team request handler.
type TeamHandler struct{}

// NewTeamHandler creates and returns a new TeamHandler.
func NewTeamHandler() *TeamHandler {
	return &TeamHandler{}
}

const teamIDSessionKey = "TeamID"

func (*TeamHandler) Authenticator(ctx context.Context, session session.Session) error {
	teamID, ok := session.Get(teamIDSessionKey).(uint)
	if !ok {
		return ctx.Error(40300, "")
	}

	team, err := db.Teams.GetByID(ctx.Request().Context(), teamID)
	if err != nil {
		if err == db.ErrTeamNotExists {
			return ctx.Error(40300, "")
		}

		log.Error("Failed to get team: %v", err)
		return ctx.ServerError()
	}

	ctx.Map(team)
	return nil
}

func (*TeamHandler) Login(ctx context.Context, session session.Session, f form.TeamLogin) error {
	team, err := db.Teams.Authenticate(ctx.Request().Context(), f.Name, f.Password)
	if err == db.ErrBadCredentials {
		return ctx.Error(40300, "bad credentials")
	} else if err != nil {
		log.Error("Failed to authenticate team: %v", err)
		return ctx.Error(50000, "")
	}

	session.Set(teamIDSessionKey, team.ID)
	return ctx.Success(session.ID())
}

func (*TeamHandler) Logout(ctx context.Context, session session.Session) error {
	session.Flush()
	return ctx.Success("")
}

func (*TeamHandler) SubmitFlag(ctx context.Context, team *db.Team) error {
	panic("implement me")
}

// Info returns the current logged in team's information.
func (*TeamHandler) Info(ctx context.Context, team *db.Team) error {
	return ctx.Success(map[string]interface{}{
		"Name":  team.Name,
		"Logo":  team.Logo,
		"Score": team.Score,
		"Rank":  team.Rank,
		"Token": team.Token,
	})
}

func (*TeamHandler) GameBoxes(ctx context.Context, team *db.Team) error {
	gameBoxes, err := db.GameBoxes.Get(ctx.Request().Context(), db.GetGameBoxesOption{
		TeamID:  team.ID,
		Visible: true,
	})
	if err != nil {
		log.Error("Failed to get team game boxes: %v", err)
		return ctx.ServerError()
	}

	type gameBox struct {
		Title       string           `json:"Title"`
		Address     string           `json:"Address"`
		Description string           `json:"Description"`
		Score       float64          `json:"Score"`
		Status      db.GameBoxStatus `json:"Status"`
	}

	gameBoxList := make([]*gameBox, 0, len(gameBoxes))
	for _, gb := range gameBoxes {
		gameBoxList = append(gameBoxList, &gameBox{
			Title:       gb.Challenge.Title,
			Address:     gb.Address,
			Description: gb.Description,
			Score:       gb.Score,
			Status:      gb.Status,
		})
	}

	return ctx.Success(gameBoxList)
}

func (*TeamHandler) Bulletins(ctx context.Context) error {
	bulletins, err := db.Bulletins.Get(ctx.Request().Context())
	if err != nil {
		log.Error("Failed to get bulletins: %v", err)
		return ctx.ServerError()
	}

	type bulletin struct {
		Title     string    `json:"Title"`
		Body      string    `json:"Body"`
		CreatedAt time.Time `json:"CreatedAt"`
	}

	bulletinList := make([]*bulletin, 0, len(bulletins))
	for _, b := range bulletins {
		bulletinList = append(bulletinList, &bulletin{
			Title:     b.Title,
			Body:      b.Body,
			CreatedAt: b.CreatedAt,
		})
	}

	return ctx.Success(bulletinList)
}
