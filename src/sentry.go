package main

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/vidar-team/Cardinal/src/conf"
)

func (s *Service) initSentry() {
	if !conf.Get().Sentry {
		return
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{IPAddress: "{{auto}}"})
	})

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://08a91604e4c9434ab6fdc6369ee577d7@o424435.ingest.sentry.io/5356242",
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			event.Tags["cardinal_version"] = VERSION
			event.Release = VERSION
			return event
		},
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	// greeting
	sentry.CaptureMessage("Hello " + VERSION)
}
