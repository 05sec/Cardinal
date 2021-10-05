// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ RanksStore = (*ranks)(nil)

// Ranks is the default instance of the RanksStore.
var Ranks RanksStore

// RanksStore is the persistent interface for ranks.
type RanksStore interface {
	// List returns the ranking list.
	List(ctx context.Context) ([]*RankItem, error)
	// VisibleChallengeTitle returns the titles of the visible challenges.
	VisibleChallengeTitle(ctx context.Context) ([]string, error)
}

// NewRanksStore returns a RanksStore instance with the given database connection.
func NewRanksStore(db *gorm.DB) RanksStore {
	return &ranks{DB: db}
}

type ranks struct {
	*gorm.DB
}

// RankItem represents a single row of the ranking list.
type RankItem struct {
	TeamID    uint
	TeamName  string
	TeamLogo  string
	Rank      uint
	Score     float64
	GameBoxes GameBoxInfoList // Ordered by challenge ID.
}

type GameBoxInfoList []*GameBoxInfo

// GameBoxInfo contains the game box info.
type GameBoxInfo struct {
	ChallengeID uint
	IsAttacked  bool
	IsDown      bool
	Score       float64 `json:",omitempty"` // Manager only
}

func (g GameBoxInfoList) Len() int           { return len(g) }
func (g GameBoxInfoList) Less(i, j int) bool { return g[i].ChallengeID < g[j].ChallengeID }
func (g GameBoxInfoList) Swap(i, j int)      { g[i], g[j] = g[j], g[i] }

type RankListOptions struct {
	ShowGameBoxScore bool
}

func (db *ranks) List(ctx context.Context) ([]*RankItem, error) {
	teamsStore := NewTeamsStore(db.DB)
	teams, err := teamsStore.Get(ctx, GetTeamsOptions{
		OrderBy: "score",
		Order:   "DESC",
	})
	if err != nil {
		return nil, errors.Wrap(err, "get teams")
	}

	rankItems := make([]*RankItem, 0, len(teams))

	gameBoxesStore := NewGameBoxesStore(db.DB)
	for _, team := range teams {
		gameBoxes, err := gameBoxesStore.Get(ctx, GetGameBoxesOption{
			TeamID:  team.ID,
			Visible: true,
		})
		if err != nil {
			return nil, errors.Wrap(err, "get team game box")
		}

		gameBoxInfo := make(GameBoxInfoList, 0, len(gameBoxes))
		for _, gameBox := range gameBoxes {
			gameBoxInfo = append(gameBoxInfo, &GameBoxInfo{
				ChallengeID: gameBox.ChallengeID,
				IsAttacked:  gameBox.Status == GameBoxStatusCaptured,
				IsDown:      gameBox.Status == GameBoxStatusDown,
				Score:       gameBox.Score,
			})
		}

		// Game box should be ordered by the challenge ID,
		// to make sure the ranking list table header can match with the score correctly.
		sort.Sort(gameBoxInfo)

		rankItems = append(rankItems, &RankItem{
			TeamID:    team.ID,
			TeamName:  team.Name,
			TeamLogo:  team.Logo,
			Rank:      team.Rank,
			Score:     team.Score,
			GameBoxes: gameBoxInfo,
		})
	}
	return rankItems, nil
}

func (db *ranks) VisibleChallengeTitle(ctx context.Context) ([]string, error) {
	var challenges []*Challenge
	if err := db.WithContext(ctx).Raw("SELECT * FROM challenges WHERE id IN " +
		"(SELECT DISTINCT game_boxes.challenge_id FROM game_boxes WHERE game_boxes.visible = TRUE AND game_boxes.deleted_at IS NULL) " +
		"AND deleted_at IS NULL ORDER BY id").Scan(&challenges).Error; err != nil {
		return nil, err
	}

	titles := make([]string, 0, len(challenges))
	for _, c := range challenges {
		titles = append(titles, c.Title)
	}
	return titles, nil
}
