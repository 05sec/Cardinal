package healthy

import (
	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/internal/db"
	"github.com/vidar-team/Cardinal/internal/utils"
	"runtime"
)

// Panel returns the system runtime status, which is used in backstage data panel.
func Panel(c *gin.Context) (int, interface{}) {
	var submitFlag int
	db.MySQL.Model(&db.AttackAction{}).Count(&submitFlag)

	var checkDown int
	db.MySQL.Model(&db.DownAction{}).Count(&checkDown)

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	return utils.MakeSuccessJSON(gin.H{
		"SubmitFlag":         submitFlag,
		"CheckDown":          checkDown,
		"NumGoroutine":       runtime.NumGoroutine(),         // Goroutine number
		"MemAllocated":       utils.FileSize(int64(m.Alloc)), // Allocated memory
		"TotalScore":         totalScore(),
		"PreviousRoundScore": previousRoundScore(),
		"Version":            utils.VERSION,
		"CommitSHA":          utils.COMMIT_SHA,
		"BuildTime":          utils.BUILD_TIME,
	})
}
