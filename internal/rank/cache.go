// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package rank

import (
	"context"

	"github.com/pkg/errors"

	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/store"
)

const (
	CacheKeyRankForTeam    = "rankForTeam"
	CacheKeyRankForManager = "rankForManager"
	CacheKeyRankTitle      = "rankTitle"
)

// ForTeam returns ranking list for team account from the cache.
// It only contains the challenges which are visible to teams.
func ForTeam() []*db.RankItem {
	rankList, ok := store.Get(CacheKeyRankForTeam)
	if !ok {
		return []*db.RankItem{}
	}
	return rankList.([]*db.RankItem)
}

// ForManager returns ranking list for team account from the cache.
// It contains all the challenges.
func ForManager() []*db.RankItem {
	rankList, ok := store.Get(CacheKeyRankForManager)
	if !ok {
		return []*db.RankItem{}
	}
	return rankList.([]*db.RankItem)
}

// Title returns the ranking list table header from the cache.
func Title() []string {
	title, ok := store.Get(CacheKeyRankTitle)
	if !ok {
		return []string{}
	}
	return title.([]string)
}

// SetTitle saves the visible challenges' headers into cache.
func SetTitle(ctx context.Context) error {
	titles, err := db.Ranks.VisibleChallengeTitle(ctx)
	if err != nil {
		return errors.Wrap(err, "get visible challenge title")
	}
	store.Set(CacheKeyRankTitle, titles)
	return nil
}

// SetRankList calculates the ranking list for teams and managers.
func SetRankList(ctx context.Context) error {
	rankList, err := db.Ranks.List(ctx)
	if err != nil {
		return errors.Wrap(err, "get rank list")
	}
	store.Set(CacheKeyRankForManager, rankList)

	// Team accounts can't get the score of the game boxes.
	for _, rankItem := range rankList {
		for _, gameBox := range rankItem.GameBoxes {
			gameBox.Score = 0
		}
	}
	store.Set(CacheKeyRankForTeam, rankList)

	return nil
}
