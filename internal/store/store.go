package store

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var store *cache.Cache

func Init() {
	c := cache.New(cache.NoExpiration, cache.DefaultExpiration)
	store = c
}

func Get(k string) (interface{}, bool) {
	return store.Get(k)
}

func Set(k string, x interface{}, d time.Duration) {
	store.Set(k, x, d)
}
