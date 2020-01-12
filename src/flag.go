package main

import (
	"github.com/jinzhu/gorm"
	"strconv"
)

type Flag struct {
	gorm.Model

	TeamID    uint
	GameBoxID uint
	Round     int
	Flag      string
}

func (s *Service) GenerateFlag() (int, interface{}) {
	var gameBoxes []GameBox
	s.Mysql.Model(&GameBox{}).Find(&gameBoxes)

	// 删库
	s.Mysql.Delete(&Flag{})

	for round := 1; round <= s.Timer.TotalRound; round++ {
		// Flag = FlagPrefix + sha1(TeamID + GameBoxID + sha1(Salt) + Round) + FlagSuffix
		for _, gameBox := range gameBoxes {
			flag := s.Conf.FlagPrefix + s.sha1Encode(strconv.Itoa(int(gameBox.TeamID))+strconv.Itoa(int(gameBox.ID))+s.sha1Encode(s.Conf.Salt)+strconv.Itoa(round)) + s.Conf.FlagSuffix
			s.Mysql.Create(&Flag{
				TeamID:    gameBox.TeamID,
				GameBoxID: gameBox.ID,
				Round:     round,
				Flag:      flag,
			})
		}
	}

	return s.makeSuccessJSON("生成 Flag 成功！")
}
