package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/src/conf"
	"github.com/vidar-team/Cardinal/src/locales"
	"github.com/vidar-team/Cardinal/src/utils"
)

// RefreshConfig put the config into the struct from database.
func (s *Service) RefreshConfig() {
	conf.Get().DynamicConfig = new(conf.DynamicConfig)
	s.Mysql.Model(conf.DynamicConfig{}).Last(conf.Get().DynamicConfig)
}

// SetConfig update the config by insert a new record into database, for we can make a config version control soon.
// Then refresh the config in struct.
func (s *Service) SetConfig(config *conf.DynamicConfig) {
	s.Mysql.Model(conf.DynamicConfig{}).Create(config)
	s.RefreshConfig()
}

func (s *Service) getConfig(c *gin.Context) (int, interface{}) {
	return utils.MakeSuccessJSON(conf.Get().DynamicConfig)
}

func (s *Service) editConfig(c *gin.Context) (int, interface{}) {
	var formData conf.DynamicConfig
	err := c.BindJSON(&formData)
	if err != nil {
		return utils.MakeErrJSON(400, 40000,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}
	s.SetConfig(&formData)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}
