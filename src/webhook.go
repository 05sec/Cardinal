package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/parnurzeal/gorequest"
	"github.com/patrickmn/go-cache"
	"github.com/thanhpk/randstr"
	"github.com/vidar-team/Cardinal/src/locales"
	"github.com/vidar-team/Cardinal/src/utils"
	"net/http"
	"strconv"
	"time"
)

const (
	ANY_HOOK         string = "any"
	NEW_ROUND_HOOK   string = "new_round"
	SUBMIT_FLAG_HOOK string = "submit_flag"
	BEGIN_HOOK       string = "game_begin"
	PAUSE_HOOK       string = "game_pause"
	END_HOOK         string = "game_end"
)

// WebHook is a gorm model for database table `webhook`, used to store the webhook.
type WebHook struct {
	gorm.Model

	URL   string
	Type  string
	Token string

	Retry   int
	Timeout int
}

func (s *Service) getWebHook(c *gin.Context) (int, interface{}) {
	var webHooks []WebHook
	s.Mysql.Model(&WebHook{}).Find(&webHooks)
	return utils.MakeSuccessJSON(webHooks)
}

func (s *Service) newWebHook(c *gin.Context) (int, interface{}) {
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
	case ANY_HOOK, NEW_ROUND_HOOK, SUBMIT_FLAG_HOOK, BEGIN_HOOK, PAUSE_HOOK, END_HOOK:
	default:
		return utils.MakeErrJSON(400, 40035,
			locales.I18n.T(c.GetString("lang"), "webhook.error_type"),
		)
	}

	if inputForm.Token == "" {
		inputForm.Token = randstr.String(32)
	}

	newWebHook := &WebHook{
		URL:     inputForm.URL,
		Type:    inputForm.Type,
		Token:   inputForm.Token,
		Retry:   inputForm.Retry,
		Timeout: inputForm.Timeout,
	}

	tx := s.Mysql.Begin()
	if tx.Create(&newWebHook).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50022,
			locales.I18n.T(c.GetString("lang"), "webhook.post_error"))
	}
	tx.Commit()

	s.refreshWebHookStore()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "webhook.post_success"))
}

func (s *Service) editWebHook(c *gin.Context) (int, interface{}) {
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

	var checkWebHook WebHook
	s.Mysql.Model(&WebHook{}).Where(&WebHook{Model: gorm.Model{ID: inputForm.ID}}).Find(&checkWebHook)
	if checkWebHook.ID == 0 {
		return utils.MakeErrJSON(404, 40405,
			locales.I18n.T(c.GetString("lang"), "webhook.not_found"),
		)
	}

	// Check type
	switch inputForm.Type {
	case ANY_HOOK, NEW_ROUND_HOOK, SUBMIT_FLAG_HOOK, BEGIN_HOOK, PAUSE_HOOK, END_HOOK:
	default:
		return utils.MakeErrJSON(400, 40035,
			locales.I18n.T(c.GetString("lang"), "webhook.error_type"),
		)
	}

	if inputForm.Token == "" {
		inputForm.Token = randstr.String(32)
	}

	editWebHook := &WebHook{
		URL:     inputForm.URL,
		Type:    inputForm.Type,
		Token:   inputForm.Token,
		Retry:   inputForm.Retry,
		Timeout: inputForm.Timeout,
	}

	tx := s.Mysql.Begin()
	if tx.Model(&WebHook{}).Where(&WebHook{Model: gorm.Model{ID: inputForm.ID}}).Update(&editWebHook).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50023,
			locales.I18n.T(c.GetString("lang"), "webhook.edit_error"))
	}
	tx.Commit()

	s.refreshWebHookStore()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "webhook.edit_success"))
}

func (s *Service) deleteWebHook(c *gin.Context) (int, interface{}) {
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

	var checkWebHook WebHook
	s.Mysql.Model(&WebHook{}).Where(&WebHook{Model: gorm.Model{ID: uint(id)}}).Find(&checkWebHook)
	if checkWebHook.ID == 0 {
		return utils.MakeErrJSON(404, 40405,
			locales.I18n.T(c.GetString("lang"), "webhook.not_found"),
		)
	}

	tx := s.Mysql.Begin()
	if tx.Where("id = ?", uint(id)).Delete(&WebHook{}).RowsAffected != 1 {
		tx.Rollback()
		return utils.MakeErrJSON(500, 50024,
			locales.I18n.T(c.GetString("lang"), "webhook.delete_error"))
	}
	tx.Commit()

	s.refreshWebHookStore()
	return utils.MakeSuccessJSON(locales.I18n.T(c.GetString("lang"), "webhook.delete_success"))
}

func (s *Service) refreshWebHookStore() {
	var webHooks []WebHook
	s.Mysql.Model(&WebHook{}).Find(&webHooks)
	s.Store.Set("webHook", webHooks, cache.NoExpiration)
}

// AddHook used to add a webhook.
func (s *Service) AddHook(webHookType string, webHookData interface{}) {
	s.sendWebHook(ANY_HOOK, webHookData)
	if webHookType != ANY_HOOK {
		s.sendWebHook(webHookType, webHookData)
	}
}

func (s *Service) sendWebHook(webHookType string, webHookData interface{}) {
	webHookStore, ok := s.Store.Get("webHook")
	if !ok {
		s.NewLog(IMPORTANT, "webhook_cache", "WebHook 缓存获取失败！")
		return
	}
	webHooks, ok := webHookStore.([]WebHook)
	if !ok {
		s.NewLog(IMPORTANT, "webhook_cache", "WebHook 缓存获取失败！")
		return
	}

	for _, v := range webHooks {
		if v.Type == webHookType {
			go func(webhook WebHook) {
				nonce := randstr.Hex(16)

				req := gorequest.New().Post(webhook.URL)
				req.Data = map[string]interface{}{
					"type": webhook.Type,
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
					s.NewLog(IMPORTANT, "webhook_cache", "WebHook 投递失败: "+webhook.URL)
				}
			}(v)
		}
	}
}
