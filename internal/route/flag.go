// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"fmt"

	log "unknwon.dev/clog/v2"

	"github.com/vidar-team/Cardinal/internal/clock"
	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/utils"
)

type FlagHandler struct{}

func NewFlagHandler() *FlagHandler {
	return &FlagHandler{}
}

func (*FlagHandler) Get(ctx context.Context) error {
	page := ctx.QueryInt("page")
	pageSize := ctx.QueryInt("pageSize")
	teamID := ctx.QueryInt("teamID")
	challengeID := ctx.QueryInt("challengeID")
	gameBoxID := ctx.QueryInt("gameBoxID")
	round := ctx.QueryInt("round")

	flags, err := db.Flags.Get(ctx.Request().Context(), db.GetFlagOptions{
		Page:        page,
		PageSize:    pageSize,
		TeamID:      uint(teamID),
		ChallengeID: uint(challengeID),
		GameBoxID:   uint(gameBoxID),
		Round:       uint(round),
	})
	if err != nil {
		log.Error("Failed to get flags: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success(flags)
}

func (*FlagHandler) BatchCreate(ctx context.Context) error {
	// TODO time analytic
	// TODO delete all the flag and regenerate in transaction.

	gameBoxes, err := db.GameBoxes.Get(ctx.Request().Context(), db.GetGameBoxesOption{})
	if err != nil {
		log.Error("Failed to get game boxes: %v", err)
		return ctx.ServerError()
	}

	flagPrefix := conf.Game.FlagPrefix
	flagSuffix := conf.Game.FlagSuffix
	salt := utils.Sha1Encode(conf.App.SecuritySalt)
	totalRound := clock.T.TotalRound

	flagMetadatas := make([]db.FlagMetadata, 0, int(totalRound)*len(gameBoxes))
	for round := uint(1); round <= totalRound; round++ {
		// Flag = FlagPrefix + hmacSha1(TeamID + | + GameBoxID + | + Round, sha1(salt)) + FlagSuffix
		for _, gameBox := range gameBoxes {
			flag := flagPrefix + utils.HmacSha1Encode(fmt.Sprintf("%d|%d|%d", gameBox.TeamID, gameBox.ID, round), salt) + flagSuffix
			flagMetadatas = append(flagMetadatas, db.FlagMetadata{
				GameBoxID: gameBox.ID,
				Round:     round,
				Value:     flag,
			})
		}
	}

	if err := db.Flags.BatchCreate(ctx.Request().Context(), db.CreateFlagOptions{
		Flags: flagMetadatas,
	}); err != nil {
		log.Error("Failed to batch create flags: %v", err)
		return ctx.ServerError()
	}

	return ctx.Success()
}
