package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/vidar-team/Cardinal/src/locales"
	"github.com/vidar-team/Cardinal/src/utils"
)

// dynamicConfig is the config which is stored in database.
// So it's a GORM model for users can edit it anytime.
type DynamicConfig struct {
	gorm.Model `json:"-"`

	Key   string
	Value string
}

// initConfig set the default value of the given key.
// Always used in installation.
func (s *Service) initConfig(key string, value string) {
	var count int
	s.Mysql.Model(&DynamicConfig{}).Where("key = ?", key).Count(&count)
	if count != 0 {
		return
	}

	tx := s.Mysql.Begin()
	if tx.Model(&DynamicConfig{}).Where("key = ?", key).Update(&DynamicConfig{
		Key:   key,
		Value: value,
	}).RowsAffected != 1 {
		tx.Rollback()
		return
	}
	tx.Commit()
}

// setConfig update the config by insert a new record into database, for we can make a config version control soon.
// Then refresh the config in struct.
func (s *Service) setConfig(key string, value string) {
	s.Mysql.FirstOrCreate(&DynamicConfig{
		Key:   key,
		Value: value,
	}, "key = ?", key)
}

// getConfig returns the config value.
func (s *Service) getConfig(key string) string {
	var config DynamicConfig
	s.Mysql.Model(&DynamicConfig{}).Where("key = ?", key).Find(&config)
	return config.Value
}

func (s *Service) setConfigHandler(c *gin.Context) (int, interface{}) {
	var inputForm struct {
		Key   string `binding:"required"`
		Value string `binding:"required"`
	}

	if err := c.BindJSON(&inputForm); err != nil {
		return utils.MakeErrJSON(400, 40046, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	s.setConfig(inputForm.Key, inputForm.Value)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "config.update_success"))
}

func (s *Service) getConfigHandler(c *gin.Context) (int, interface{}) {
	var inputForm struct {
		Key string `binding:"required"`
	}

	if err := c.BindJSON(&inputForm); err != nil {
		return utils.MakeErrJSON(400, 40046, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	value := s.getConfig(inputForm.Key)
	return utils.MakeSuccessJSON(value)
}
