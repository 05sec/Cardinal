package main

import "github.com/patrickmn/go-cache"

func (s *Service) initStore() {
	c := cache.New(cache.NoExpiration, cache.DefaultExpiration)
	s.Store = c

	s.refreshWebHookStore()
}
