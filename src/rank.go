package main

import "github.com/patrickmn/go-cache"

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

// 计算并存储总排行榜
func (s *Service) GenerateRankList() {
	var rankList []*RankItem
	var teams []Team
	s.Mysql.Model(&Team{}).Order("score").Find(&teams) // 根据队伍总分排序
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
	if len(rankList) != 0 {
		// 获得开放的 Challenge
		var visibleChallengeID []uint
		s.Mysql.Model(&GameBox{}).Where(&GameBox{TeamID: rankList[0].TeamID, Visible: true}).Order("challenge_id").Pluck("id", &visibleChallengeID)
		var visibleChallengeTitle []string
		s.Mysql.Model(&Challenge{}).Where("id in (?)", visibleChallengeID).Order("id").Pluck("title", &visibleChallengeTitle)
		s.Store.Set("rankListTitle", visibleChallengeTitle, cache.NoExpiration) // 存储 Challenge Title
	}
	s.Store.Set("rankList", rankList, cache.NoExpiration)
}
