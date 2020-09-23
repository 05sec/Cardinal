package misc

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"
	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/utils"
	"github.com/vidar-team/Cardinal/locales"
)

const GITHUB_RELEASE_API = "https://api.github.com/repos/vidar-team/Cardinal/releases/latest"

func CheckVersion() {
	// Check Cardinal version.
	resp, body, _ := gorequest.New().Get(GITHUB_RELEASE_API).End()
	if resp != nil && resp.StatusCode == 200 {
		type releaseApiJson struct {
			Name        string `json:"name"`
			NodeID      string `json:"node_id"`
			PublishedAt string `json:"published_at"`
			TagName     string `json:"tag_name"`
		}

		var releaseData releaseApiJson
		err := json.Unmarshal([]byte(body), &releaseData)
		if err == nil {
			// Compare version.
			if !utils.CompareVersion(utils.VERSION, releaseData.TagName) {
				log.Println(locales.I18n.T(conf.Get().SystemLanguage, "misc.version_out_of_date", gin.H{
					"currentVersion": utils.VERSION,
					"latestVersion":  releaseData.TagName,
				}))
			}
		}
	}
}
