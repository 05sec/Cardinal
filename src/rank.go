package main

import (
	"fmt"
	"github.com/patrickmn/go-cache"
)

// 排行榜
type RankItem struct {
	TeamID        uint
	TeamName      string
	TeamLogo      string
	Score         float64
	GameBoxStatus interface{} // 按照 ChallengeID 顺序
}

// 管理端靶机信息
type GameBoxInfo struct {
	Score      float64
	IsAttacked bool
	IsDown     bool
}

// 选手端靶机信息
type GameBoxStatus struct {
	IsAttacked bool
	IsDown     bool
}

// 获取选手端总排行榜内容
func (s *Service) GetRankList() []*RankItem {
	rankList, ok := s.Store.Get("rankList")
	if !ok {
		return []*RankItem{}
	}
	return rankList.([]*RankItem)
}

func (s *Service) GetManagerRankList() []*RankItem {
	rankList, ok := s.Store.Get("rankManagerList")
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

	s.NewLog(NORMAL, "system", fmt.Sprintf("更新排行榜标题成功"))
}

// 计算并存储总排行榜
func (s *Service) SetRankList() {
	var rankList []*RankItem
	var managerRankList []*RankItem

	var teams []Team
	s.Mysql.Model(&Team{}).Order("score DESC").Find(&teams) // 根据队伍总分排序
	for _, team := range teams {
		var gameboxes []GameBox
		s.Mysql.Model(&GameBox{}).Where(&GameBox{TeamID: team.ID, Visible: true}).Order("challenge_id").Find(&gameboxes) // 排序保证题目顺序一致
		var gameBoxInfo []*GameBoxInfo                                                                                   // 管理端靶机信息
		var gameBoxStatuses []*GameBoxStatus                                                                             // 当前队伍所有靶机状态

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

	// 存储总排行榜
	s.Store.Set("rankList", rankList, cache.NoExpiration)
	// 存储管理员排行榜
	s.Store.Set("rankManagerList", managerRankList, cache.NoExpiration)
	s.NewLog(NORMAL, "system", fmt.Sprintf("更新总排行榜成功！"))
}
