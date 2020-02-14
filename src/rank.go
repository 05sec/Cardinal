package main

import (
	"github.com/patrickmn/go-cache"
)

// 排行榜
type RankItem struct {
	TeamID        uint
	TeamName      string
	TeamLogo      string
	Score         float64
	GameBoxStatus []*GameBoxStatus // 按照 ChallengeID 顺序
}

type GameBoxStatus struct {
	IsAttacked bool
	IsDown     bool
}

// 获取总排行榜内容
func (s *Service) GetRankList() []*RankItem {
	rankList, ok := s.Store.Get("rankList")
	if !ok {
		return []*RankItem{}
	}
	return rankList.([]*RankItem)
}

// 获取总排行榜标题
func (s *Service) GetRankListTitle() []string {
	rankListTitle, ok := s.Store.Get("rankListTitle")
	if !ok {
		return []string{}
	}
	return rankListTitle.([]string)
}

// 存储总排行榜标题
func (s *Service) SetRankListTitle() {
	var result []struct {
		Title string `gorm:"Column:Title"`
	}
	s.Mysql.Raw("SELECT `challenges`.`Title` FROM `challenges` WHERE `challenges`.`id` IN " +
		"(SELECT DISTINCT challenge_id FROM `game_boxes` WHERE `visible` = 1 AND `deleted_at` IS NULL) " + // 获得开放的题目 ID 并去重
		"AND `deleted_at` IS NULL ORDER BY `challenges`.`id`").Scan(&result)

	visibleChallengeTitle := make([]string, len(result))
	for index, res := range result {
		visibleChallengeTitle[index] = res.Title
	}
	s.Store.Set("rankListTitle", visibleChallengeTitle, cache.NoExpiration) // 存储 Challenge Title
}

// 计算并存储总排行榜
func (s *Service) SetRankList() {
	var rankList []*RankItem
	var teams []Team
	s.Mysql.Model(&Team{}).Order("score DESC").Find(&teams) // 根据队伍总分排序
	for _, team := range teams {
		var gameboxes []GameBox
		s.Mysql.Model(&GameBox{}).Where(&GameBox{TeamID: team.ID, Visible: true}).Order("challenge_id").Find(&gameboxes) // 排序保证题目顺序一致
		var gameBoxStatuses []*GameBoxStatus                                                                             // 当前队伍所有靶机状态
		for _, gamebox := range gameboxes {
			gameBoxStatuses = append(gameBoxStatuses, &GameBoxStatus{
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
	}

	// 存储总排行榜
	s.Store.Set("rankList", rankList, cache.NoExpiration)
}
