package webhook

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/parnurzeal/gorequest"
	"github.com/patrickmn/go-cache"
	"github.com/thanhpk/randstr"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/locales"
	"github.com/vidar-team/Cardinal/internal/logger"
	"github.com/vidar-team/Cardinal/internal/store"
	"github.com/vidar-team/Cardinal/internal/utils"
)

const (
	ANY_HOOK         string = "any"
	NEW_ROUND_HOOK   string = "new_round"
	SUBMIT_FLAG_HOOK string = "submit_flag"
	CHECK_DOWN_HOOK  string = "check_down"
	BEGIN_HOOK       string = "game_begin"
	PAUSE_HOOK       string = "game_pause"
	END_HOOK         string = "game_end"
)

func GetWebHook(c *gin.Context) (int, interface{}) {
	var webHooks []db.WebHook
	db.MySQL.Model(&db.WebHook{}).Find(&webHooks)
	return utils.MakeSuccessJSON(webHooks)
}

func NewWebHook(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		URL   string `binding:"required"`
		Type  string `binding:"required"`
		Token string // If the token is empty, generate one automatically.

		Retry   int
		Timeout int
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40033,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	// Check type
	switch inputForm.Type {
	case ANY_HOOK, NEW_ROUND_HOOK, SUBMIT_FLAG_HOOK, CHECK_DOWN_HOOK, BEGIN_HOOK, PAUSE_HOOK, END_HOOK:
	default:
		return utils.MakeErrJSON(400, 40035,
			locales.I18n.T(c.GetString("lang"), "webhook.error_type"),
		)
	}

	if inputForm.Token == "" {
		inputForm.Token = randstr.String(32)
	}

	newWebHook := &db.WebHook{
		URL:     inputForm.URL,
		Type:    inputForm.Type,
		Token:   inputForm.Token,
		Retry:   inputForm.Retry,
		Timeout: inputForm.Timeout,
	}

	tx := db.MySQL.Begin()
	if tx.Create(&newWebHook).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50022,
			locales.I18n.T(c.GetString("lang"), "webhook.post_error"))
	}
	tx.Commit()

	RefreshWebHookStore()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "webhook.post_success"))
}

func EditWebHook(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID    uint   `binding:"required"`
		URL   string `binding:"required"`
		Type  string `binding:"required"`
		Token string

		Retry   int
		Timeout int
	}

	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return utils.MakeErrJSON(400, 40033,
			locales.I18n.T(c.GetString("lang"), "general.error_payload"),
		)
	}

	var checkWebHook db.WebHook
	db.MySQL.Model(&db.WebHook{}).Where(&db.WebHook{Model: gorm.Model{ID: inputForm.ID}}).Find(&checkWebHook)
	if checkWebHook.ID == 0 {
		return utils.MakeErrJSON(404, 40405,
			locales.I18n.T(c.GetString("lang"), "webhook.not_found"),
		)
	}

	// Check type
	switch inputForm.Type {
	case ANY_HOOK, NEW_ROUND_HOOK, SUBMIT_FLAG_HOOK, CHECK_DOWN_HOOK, BEGIN_HOOK, PAUSE_HOOK, END_HOOK:
	default:
		return utils.MakeErrJSON(400, 40035,
			locales.I18n.T(c.GetString("lang"), "webhook.error_type"),
		)
	}

	if inputForm.Token == "" {
		inputForm.Token = randstr.String(32)
	}

	editWebHook := &db.WebHook{
		URL:     inputForm.URL,
		Type:    inputForm.Type,
		Token:   inputForm.Token,
		Retry:   inputForm.Retry,
		Timeout: inputForm.Timeout,
	}

	tx := db.MySQL.Begin()
	if tx.Model(&db.WebHook{}).Where(&db.WebHook{Model: gorm.Model{ID: inputForm.ID}}).Update(&editWebHook).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50023,
			locales.I18n.T(c.GetString("lang"), "webhook.edit_error"))
	}
	tx.Commit()

	RefreshWebHookStore()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "webhook.edit_success"))
}

func DeleteWebHook(c *gin.Context) (int, interface{}) {
	idStr, ok := c.GetQuery("id")
	if !ok {
		return utils.MakeErrJSON(400, 40035,
			locales.I18n.T(c.GetString("lang"), "general.error_query"),
		)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return utils.MakeErrJSON(400, 40035,
			locales.I18n.T(c.GetString("lang"), "general.must_be_number", gin.H{"key": "id"}),
		)
	}

	var checkWebHook db.WebHook
	db.MySQL.Model(&db.WebHook{}).Where(&db.WebHook{Model: gorm.Model{ID: uint(id)}}).Find(&checkWebHook)
	if checkWebHook.ID == 0 {
		return utils.MakeErrJSON(404, 40405,
			locales.I18n.T(c.GetString("lang"), "webhook.not_found"),
		)
	}

	tx := db.MySQL.Begin()
	if tx.Where("id = ?", uint(id)).Delete(&db.WebHook{}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50024,
			locales.I18n.T(c.GetString("lang"), "webhook.delete_error"))
	}
	tx.Commit()

	RefreshWebHookStore()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "webhook.delete_success"))
}

func RefreshWebHookStore() {
	var webHooks []db.WebHook
	db.MySQL.Model(&db.WebHook{}).Find(&webHooks)
	store.Set("webHook", webHooks, cache.NoExpiration)
}

// Add used to add a webhook.
func Add(webHookType string, webHookData interface{}) {
	sendWebHook(webHookType, webHookData)
}

func sendWebHook(webHookType string, webHookData interface{}) {
	webHookStore, ok := store.Get("webHook")
	if !ok {
		logger.New(logger.IMPORTANT, "webhook_cache", "WebHook 缓存获取失败！")
		return
	}
	webHooks, ok := webHookStore.([]db.WebHook)
	if !ok {
		logger.New(logger.IMPORTANT, "webhook_cache", "WebHook 缓存获取失败！")
		return
	}

	for _, v := range webHooks {
		if v.Type == webHookType || v.Type == ANY_HOOK {
			go func(webhook db.WebHook) {
				nonce := randstr.Hex(16)

				req := gorequest.New().Post(webhook.URL)
				req.Data = map[string]interface{}{
					"type": webHookType,
					"data": webHookData,
					// TODO: Here is not secure. The data should be added in the signature.
					"nonce":     nonce,
					"signature": utils.HmacSha1Encode(nonce, webhook.Token),
				}
				req.Retry(webhook.Retry, time.Duration(webhook.Timeout)*time.Second,
					http.StatusInternalServerError,
					http.StatusBadGateway,
					http.StatusBadRequest,
				)
				resp, _, _ := req.End()

				if resp == nil || resp.StatusCode != 200 {
					logger.New(logger.IMPORTANT, "webhook_cache", "WebHook 投递失败: "+webhook.URL)
				}
			}(v)
		}
	}
}
