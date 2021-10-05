// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/vidar-team/Cardinal/internal/conf"
)

var _ ScoresStore = (*scores)(nil)

// Scores is the default instance of the ScoresStore.
var Scores ScoresStore

// ScoresStore is the persistent interface for scores.
type ScoresStore interface {
	Calculate(ctx context.Context, round uint) error
	RefreshAttackScore(ctx context.Context, round uint, replaces ...bool) error
	RefreshCheckScore(ctx context.Context, round uint, replaces ...bool) error
	RefreshGameBoxScore(ctx context.Context) error
	RefreshTeamScore(ctx context.Context) error
}

// NewScoresStore returns a ScoresStore instance with the given database connection.
func NewScoresStore(db *gorm.DB) ScoresStore {
	return &scores{DB: db}
}

type scores struct {
	*gorm.DB
}

// Calculate calculates the score until the given round.
func (db *scores) Calculate(ctx context.Context, round uint) error {
	if err := db.RefreshAttackScore(ctx, round); err != nil {
		return errors.Wrap(err, "refresh attack score")
	}

	if err := db.RefreshCheckScore(ctx, round); err != nil {
		return errors.Wrap(err, "refresh check score")
	}

	if err := db.RefreshGameBoxScore(ctx); err != nil {
		return errors.Wrap(err, "refresh game box score")
	}

	if err := db.RefreshTeamScore(ctx); err != nil {
		return errors.Wrap(err, "refresh team score")
	}

	// TODO: Logger

	// TODO: health check

	return nil
}

func (db *scores) RefreshAttackScore(ctx context.Context, round uint, replaces ...bool) error {
	replace := len(replaces) != 0 && replaces[0]

	gameBoxesStore := NewGameBoxesStore(db.DB)
	gameBoxes, err := gameBoxesStore.Get(ctx, GetGameBoxesOption{})
	if err != nil {
		return errors.Wrap(err, "get game boxes")
	}

	actionsStore := NewActionsStore(db.DB)

	for _, gameBox := range gameBoxes {
		// [-] Been attacked score
		beenAttackedActions, err := actionsStore.Get(ctx, GetActionOptions{
			Type:      ActionTypeBeenAttack,
			GameBoxID: gameBox.ID,
			Round:     round,
		})
		if err != nil {
			return errors.Wrap(err, "get been attacked actions")
		}

		if len(beenAttackedActions) == 0 {
			continue
		}
		beenAttackAction := beenAttackedActions[0]
		if err := actionsStore.SetScore(ctx, SetActionScoreOptions{
			ActionID: beenAttackAction.ID,
			Score:    -float64(conf.Game.AttackScore),
			Replace:  replace,
		}); err != nil {
			return errors.Wrap(err, "set been attacked score")
		}

		// [+] Attacked score
		attackActions, err := actionsStore.Get(ctx, GetActionOptions{
			Type:      ActionTypeAttack,
			GameBoxID: gameBox.ID,
			Round:     round,
		})
		if err != nil {
			return errors.Wrap(err, "get attack actions")
		}
		if len(attackActions) == 0 {
			return errors.Errorf("unexpected count of attack actions: %d", len(attackActions))
		}

		score := float64(conf.Game.AttackScore) / float64(len(attackActions))
		for _, action := range attackActions {
			if err := actionsStore.SetScore(ctx, SetActionScoreOptions{
				ActionID: action.ID,
				Score:    score,
				Replace:  replace,
			}); err != nil {
				return errors.Wrap(err, "set attack score")
			}
		}
	}

	return nil
}

func (db *scores) RefreshCheckScore(ctx context.Context, round uint, replaces ...bool) error {
	replace := len(replaces) != 0 && replaces[0]

	challengesStore := NewChallengesStore(db.DB)
	gameBoxesStore := NewGameBoxesStore(db.DB)
	actionsStore := NewActionsStore(db.DB)

	challenges, err := challengesStore.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "get challenges")
	}

	for _, challenge := range challenges {
		// Get the game boxes of the challenge.
		gameBoxes, err := gameBoxesStore.Get(ctx, GetGameBoxesOption{
			ChallengeID: challenge.ID,
		})
		if err != nil {
			return errors.Wrap(err, "get game boxes")
		}

		// Skip the invisible challenge.
		if len(gameBoxes) == 0 || !gameBoxes[0].Visible {
			continue
		}

		// Get the game box check down actions of the challenge.
		checkDownActions, err := actionsStore.Get(ctx, GetActionOptions{
			Type:        ActionTypeCheckDown,
			ChallengeID: challenge.ID,
			Round:       round,
		})
		if err != nil {
			return errors.Wrap(err, "get challenge check down boxes")
		}

		// We need save the check down game box IDs of this challenge into a map,
		// for we can get the service online game boxes when traversal all the game boxes.
		checkDownGameBoxIDs := make(map[uint]struct{}, len(checkDownActions))
		// [-] Been checked down
		for _, action := range checkDownActions {
			checkDownGameBoxIDs[action.GameBoxID] = struct{}{}

			if err := actionsStore.SetScore(ctx, SetActionScoreOptions{
				ActionID: action.ID,
				Score:    -float64(conf.Game.CheckDownScore),
				Replace:  replace,
			}); err != nil {
				return errors.Wrap(err, "set check down score")
			}
		}

		// [+] Service online
		// Remove service online actions in given round first.
		if err := actionsStore.Delete(ctx, DeleteActionOptions{
			Type:  ActionTypeServiceOnline,
			Round: round,
		}); err != nil {
			return errors.Wrap(err, "delete previous service online actions")
		}

		serviceOnlineGameBoxCount := len(gameBoxes) - len(checkDownActions)
		score := float64(conf.Game.CheckDownScore*len(checkDownActions)) / float64(serviceOnlineGameBoxCount)

		for _, gameBox := range gameBoxes {
			if _, ok := checkDownGameBoxIDs[gameBox.ID]; ok {
				continue
			}

			action, err := actionsStore.Create(ctx, CreateActionOptions{
				Type:      ActionTypeServiceOnline,
				GameBoxID: gameBox.ID,
				Round:     round,
			})
			if err != nil {
				return errors.Wrap(err, "create service online action")
			}

			if err := actionsStore.SetScore(ctx, SetActionScoreOptions{
				ActionID: action.ID,
				Score:    score,
				Replace:  true,
			}); err != nil {
				return errors.Wrap(err, "set service online action score")
			}
		}
	}

	return nil
}

func (db *scores) RefreshGameBoxScore(ctx context.Context) error {
	gameBoxesStore := NewGameBoxesStore(db.DB)
	actionsStore := NewActionsStore(db.DB)

	gameBoxes, err := gameBoxesStore.Get(ctx, GetGameBoxesOption{
		Visible: true,
	})
	if err != nil {
		return errors.Wrap(err, "get game boxes")
	}

	for _, gameBox := range gameBoxes {
		score, err := actionsStore.CountScore(ctx, CountActionScoreOptions{
			GameBoxID: gameBox.ID,
		})
		if err != nil {
			return errors.Wrap(err, "count game box score")
		}

		if err := gameBoxesStore.SetScore(ctx, gameBox.ID, score); err != nil {
			return errors.Wrap(err, "set game box score")
		}
	}

	return nil
}

func (db *scores) RefreshTeamScore(ctx context.Context) error {
	teamsStore := NewTeamsStore(db.DB)
	teams, err := teamsStore.Get(ctx, GetTeamsOptions{})
	if err != nil {
		return errors.Wrap(err, "get teams")
	}

	gameBoxesStore := NewGameBoxesStore(db.DB)
	for _, team := range teams {
		score, err := gameBoxesStore.CountScore(ctx, GameBoxCountScoreOptions{
			TeamID:  team.ID,
			Visible: true,
		})
		if err != nil {
			return errors.Wrap(err, "get game box score")
		}

		if err := teamsStore.SetScore(ctx, team.ID, score); err != nil {
			return errors.Wrap(err, "set team score")
		}
	}

	return nil
}
