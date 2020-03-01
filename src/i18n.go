package main

import (
	"github.com/gin-gonic/gin"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
	"golang.org/x/text/language"
)

func (s *Service) initI18n() {
	I18n := i18n.New(
		yaml.New("./locales"),
	)

	s.I18n = I18n
}

// I18nMiddleware is an i18n middleware. Get client language from Accept-Language header.
func (s *Service) I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptLanguages := c.GetHeader("Accept-Language")
		languages, _, err := language.ParseAcceptLanguage(acceptLanguages)
		if err != nil || len(languages) == 0 {
			c.Set("lang", "")
			c.Next()
		}

		// Only get the first language, ignore the rest.
		c.Set("lang", languages[0].String())
		c.Next()
	}
}
