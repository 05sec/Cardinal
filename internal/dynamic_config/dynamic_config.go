package dynamic_config

import (
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/utils"
)

func Init() {
	db.MySQL.Model(&db.DynamicConfig{})

	initConfig(utils.DATBASE_VERSION, db.VERSION, utils.STRING)
	initConfig(utils.TITLE_CONF, "HCTF", utils.STRING)
	initConfig(utils.FLAG_PREFIX_CONF, "hctf{", utils.STRING)
	initConfig(utils.FLAG_SUFFIX_CONF, "}", utils.STRING)
	initConfig(utils.ANIMATE_ASTEROID, utils.BOOLEAN_FALSE, utils.BOOLEAN)
	initConfig(utils.SHOW_OTHERS_GAMEBOX, utils.BOOLEAN_FALSE, utils.BOOLEAN)
}

// initConfig set the default value of the given key.
// Always used in installation.
func initConfig(key string, value string, kind int8) {
	db.MySQL.Model(&db.DynamicConfig{}).FirstOrCreate(&db.DynamicConfig{
		Key:   key,
		Value: value,
		Kind:  kind,
	}, "`key` = ?", key)
}

// Set update the config by insert a new record into database, for we can make a config version control soon.
// Then refresh the config in struct.
func Set(key string, value string) {
	db.MySQL.Model(&db.DynamicConfig{}).Where("`key` = ?", key).Update(&db.DynamicConfig{
		Key:   key,
		Value: value,
	})
}

// Get returns the config value.
func Get(key string) string {
	var config db.DynamicConfig
	db.MySQL.Model(&db.DynamicConfig{}).Where("`key` = ?", key).Find(&config)
	return config.Value
}

// SetConfig is the HTTP handler used to set the config value.
func SetConfig(c *gin.Context) (int, interface{}) {
	var inputForm []struct {
		Key   string `binding:"required"`
		Value string `binding:"required"`
	}

	if err := c.BindJSON(&inputForm); err != nil {
		return utils.MakeErrJSON(400, 40046, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}

	for _, config := range inputForm {
		Set(config.Key, config.Value)
	}
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "config.update_success"))
}

// GetConfig is the HTTP handler used to return the config value of the given key.
func GetConfig(c *gin.Context) (int, interface{}) {
	var inputForm struct {
		Key string `binding:"required"`
	}

	if err := c.BindJSON(&inputForm); err != nil {
		return utils.MakeErrJSON(400, 40046, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	value := Get(inputForm.Key)
	return utils.MakeSuccessJSON(value)
}

// GetAllConfig is the HTTP handler used to return the all the configs.
func GetAllConfig(c *gin.Context) (int, interface{}) {
	var config []db.DynamicConfig
	db.MySQL.Model(&db.DynamicConfig{}).Find(&config)
	return utils.MakeSuccessJSON(config)
}
