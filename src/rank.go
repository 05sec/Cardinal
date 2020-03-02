package main

import (
	"github.com/patrickmn/go-cache"
)

// RankItem is used to create the ranking list.
type RankItem struct {
	TeamID        uint
	TeamName      string
	TeamLogo      string
	Score         float64
	GameBoxStatus interface{} // Ordered by challenge ID.
}

// GameBoxInfo contains the gamebox info which for manager.
// Manager can get the gamebox's score.
type GameBoxInfo struct {
	Score      float64
	IsAttacked bool
	IsDown     bool
}

// GameBoxStatus contains the gamebox info which for team.
type GameBoxStatus struct {
	IsAttacked bool
	IsDown     bool
}

// GetRankList returns the ranking list data for team from the cache.
func (s *Service) GetRankList() []*RankItem {
	rankList, ok := s.Store.Get("rankList")
	if !ok {
		return []*RankItem{}
	}
	return rankList.([]*RankItem)
}

// GetManagerRankList returns the ranking list data for manager from the cache.
func (s *Service) GetManagerRankList() []*RankItem {
	rankList, ok := s.Store.Get("rankManagerList")
	if !ok {
		return []*RankItem{}
	}
	return rankList.([]*RankItem)
}

// GetRankListTitle returns the ranking list table header from the cache.
func (s *Service) GetRankListTitle() []string {
	rankListTitle, ok := s.Store.Get("rankListTitle")
	if !ok {
		return []string{}
	}
	return rankListTitle.([]string)
}

// SetRankListTitle will save the visible challenges' headers into cache.
func (s *Service) SetRankListTitle() {
	var result []struct {
		Title string `gorm:"Column:Title"`
	}
	s.Mysql.Raw("SELECT `challenges`.`Title` FROM `challenges` WHERE `challenges`.`id` IN " +
		"(SELECT DISTINCT challenge_id FROM `game_boxes` WHERE `visible` = 1 AND `deleted_at` IS NULL) " + // DISTINCT get all the visible challenge IDs and remove duplicate data
		"AND `deleted_at` IS NULL ORDER BY `challenges`.`id`").Scan(&result)

	visibleChallengeTitle := make([]string, len(result))
	for index, res := range result {
		visibleChallengeTitle[index] = res.Title
	}
	s.Store.Set("rankListTitle", visibleChallengeTitle, cache.NoExpiration) // Save challenge title into cache.

	s.NewLog(NORMAL, "system", string(s.I18n.T(s.Conf.Base.SystemLanguage, "log.rank_list_success")))
}

// SetRankList will calculate the ranking list.
func (s *Service) SetRankList() {
	var rankList []*RankItem
	var managerRankList []*RankItem

	var teams []Team
	s.Mysql.Model(&Team{}).Order("score DESC").Find(&teams) // Ordered by the team score.
	for _, team := range teams {
		var gameboxes []GameBox
		// Get the challenge data ordered by the challenge ID, to make sure the table header can match with the score correctly.
		s.Mysql.Model(&GameBox{}).Where(&GameBox{TeamID: team.ID, Visible: true}).Order("challenge_id").Find(&gameboxes)
		var gameBoxInfo []*GameBoxInfo       // Gamebox info for manager.
		var gameBoxStatuses []*GameBoxStatus // Gamebox info for users and public.

		for _, gamebox := range gameboxes {
			gameBoxStatuses = append(gameBoxStatuses, &GameBoxStatus{
				IsAttacked: gamebox.IsAttacked,
				IsDown:     gamebox.IsDown,
			})

			gameBoxInfo = append(gameBoxInfo, &GameBoxInfo{
				Score:      gamebox.Score,
				IsAttacked: gamebox.IsAttacked,
				IsDown:     gamebox.IsDown,
			})
		}

		rankList = append(rankList, &RankItem{
			TeamID:        team.ID,
			TeamName:      team.Name,
			TeamLogo:      team.Logo,
			Score:         team.Score,
			GameBoxStatus: gameBoxStatuses,
		})
		managerRankList = append(managerRankList, &RankItem{
			TeamID:        team.ID,
			TeamName:      team.Name,
			TeamLogo:      team.Logo,
			Score:         team.Score,
			GameBoxStatus: gameBoxInfo,
		})
	}

	// Save the ranking list for public into cache.
	s.Store.Set("rankList", rankList, cache.NoExpiration)
	// Save the ranking list for manager into cache.
	s.Store.Set("rankManagerList", managerRankList, cache.NoExpiration)
}
