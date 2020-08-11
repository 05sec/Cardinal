package store

import (
	"github.com/patrickmn/go-cache"
	"github.com/vidar-team/Cardinal/internal/misc/webhook"
	"time"
)

var store *cache.Cache

func Init() {
	c := cache.New(cache.NoExpiration, cache.DefaultExpiration)
	store = c

	webhook.RefreshWebHookStore()
}

func Get(k string) (interface{}, bool) {
	return store.Get(k)
}

func Set(k string, x interface{}, d time.Duration) {
	store.Set(k, x, d)
}
