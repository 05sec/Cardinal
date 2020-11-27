package asteroid

import (
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/utils"
)

func GetAsteroidStatus(c *gin.Context) (int, interface{}) {
	return utils.MakeSuccessJSON(gin.H{
		// TODO
	})
}

func Attack(c *gin.Context) (int, interface{}) {
	var attackData struct {
		From int `binding:"required"`
		To   int `binding:"required"`
	}
	if err := c.BindJSON(&attackData); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	sendAttack(attackData.From, attackData.To)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func Rank(c *gin.Context) (int, interface{}) {
	sendRank()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func Status(c *gin.Context) (int, interface{}) {
	var status struct {
		Id     int    `binding:"required"`
		Status string `binding:"required"`
	}
	if err := c.BindJSON(&status); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	if status.Status != "down" && status.Status != "attacked" {
		return utils.MakeErrJSON(400, 40039, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	sendStatus(status.Id, status.Status)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func Round(c *gin.Context) (int, interface{}) {
	var round struct {
		Round int `binding:"required"`
	}
	if err := c.BindJSON(&round); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	sendRound(round.Round)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func EasterEgg(c *gin.Context) (int, interface{}) {
	sendEasterEgg()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func Time(c *gin.Context) (int, interface{}) {
	var time struct {
		Time int `binding:"required"`
	}
	if err := c.BindJSON(&time); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	sendTime(time.Time)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func Clear(c *gin.Context) (int, interface{}) {
	var clear struct {
		Id int `binding:"required"`
	}
	if err := c.BindJSON(&clear); err != nil {
		return utils.MakeErrJSON(400, 40038, locales.I18n.T(c.GetString("lang"), "general.error_payload"))
	}
	sendClear(clear.Id)
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}

func ClearAll(c *gin.Context) (int, interface{}) {
	sendClearAll()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "general.success"))
}
