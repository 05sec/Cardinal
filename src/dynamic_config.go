package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strconv"
)

// DynamicConfig is the config which is stored in database.
// So it's a GORM model for users can edit it anytime.
type DynamicConfig struct {
	gorm.Model
	Key   string
	Value string
}

func (s *Service) initDynamicConfig() {
	if (!s.Mysql.HasTable(&DynamicConfig{})) {
		s.Mysql.AutoMigrate(&DynamicConfig{})

		// Set default data
		s.Set("title", "HCTF")
		s.Set("asteroid_enabled", false)
	}
}

func (s *Service) Set(key string, value interface{}) {
	v := fmt.Sprintf("%v", value)
	var count int
	s.Mysql.Model(&DynamicConfig{}).Where(&DynamicConfig{Key: key}).Count(&count)
	if count == 0 {
		s.Mysql.Create(&DynamicConfig{
			Key:   key,
			Value: v,
		})
	} else {
		s.Mysql.Model(&DynamicConfig{}).Where("key = ?", key).Updates(map[string]interface{}{"value": v})
	}
}

func (s *Service) GetString(key string) string {
	var conf DynamicConfig
	s.Mysql.Model(&DynamicConfig{}).Where(&DynamicConfig{Key: key}).Find(&conf)
	return conf.Value
}

func (s *Service) GetBool(key string) bool {
	var conf DynamicConfig
	s.Mysql.Model(&DynamicConfig{}).Where(&DynamicConfig{Key: key}).Find(&conf)
	if value, ok := strconv.ParseBool(conf.Value); ok == nil {
		return value
	}
	return false
}

func (s *Service) GetInt(key string) int {
	var conf DynamicConfig
	s.Mysql.Model(&DynamicConfig{}).Where(&DynamicConfig{Key: key}).Find(&conf)
	if value, ok := strconv.ParseInt(conf.Value, 10, 32); ok == nil {
		return int(value)
	}
	return 0
}
