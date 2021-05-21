package healthy

import (
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/vidar-team/Cardinal/internal/dbold"
	"github.com/vidar-team/Cardinal/internal/utils"
)

// Panel returns the system runtime status, which is used in backstage data panel.
func Panel(c *gin.Context) (int, interface{}) {
	var submitFlag int
	dbold.MySQL.Model(&dbold.AttackAction{}).Count(&submitFlag)

	var checkDown int
	dbold.MySQL.Model(&dbold.DownAction{}).Count(&checkDown)

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	return utils.MakeSuccessJSON(gin.H{
		"SubmitFlag":         submitFlag,
		"CheckDown":          checkDown,
		"NumGoroutine":       runtime.NumGoroutine(),         // Goroutine number
		"MemAllocated":       utils.FileSize(int64(m.Alloc)), // Allocated memory
		"TotalScore":         TotalScore(),
		"PreviousRoundScore": PreviousRoundScore(),
		"Version":            utils.VERSION,
		"CommitSHA":          utils.COMMIT_SHA,
		"BuildTime":          utils.BUILD_TIME,
	})
}
