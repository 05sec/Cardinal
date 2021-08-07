package locales

import (
	"github.com/gin-gonic/gin"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
	"golang.org/x/text/language"

	"github.com/vidar-team/Cardinal/internal/conf"
)

// I18n is the i18n constant.
var I18n *i18n.I18n

func init() {
	I18n = i18n.New(
		yaml.New("./locales"),
	)
}

// T returns the translation of the given key in the default language.
func T(key string, args ...interface{}) string {
	return string(I18n.T(conf.App.Language, key, args...))
}

// Middleware is an i18n middleware. Get client language from Accept-Language header.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptLanguages := c.GetHeader("Accept-Language")
		languages, _, err := language.ParseAcceptLanguage(acceptLanguages)
		if err != nil || len(languages) == 0 {
			// Set en-US as default language.
			c.Set("lang", "en-US")
			c.Next()
			return
		}

		// Only get the first language, ignore the rest.
		c.Set("lang", languages[0].String())
		c.Next()
	}
}
