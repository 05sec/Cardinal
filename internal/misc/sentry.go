package misc

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/vidar-team/Cardinal/conf"
	"github.com/vidar-team/Cardinal/internal/utils"
)

const sentryDSN = "https://08a91604e4c9434ab6fdc6369ee577d7@o424435.ingest.sentry.io/5356242"

func Sentry() {
	cardinalVersion := utils.VERSION
	cardinalCommitSHA := utils.COMMIT_SHA

	if !conf.Get().Sentry {
		return
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{IPAddress: "{{auto}}"})
	})

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: sentryDSN,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			event.Tags["cardinal_version"] = cardinalVersion
			event.Release = cardinalCommitSHA
			return event
		},
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	// greeting
	sentry.CaptureMessage("Hello " + cardinalVersion)
}
