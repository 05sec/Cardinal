// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package conf

import (
	"os"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	log "unknwon.dev/clog/v2"
)

func init() {
	err := log.NewConsole()
	if err != nil {
		panic("init console logger: " + err.Error())
	}
}

func Init(customConf string) error {
	if customConf == "" {
		customConf = "./conf/Cardinal.toml"
	}

	config, err := toml.LoadFile(customConf)
	if err != nil {
		return errors.Wrap(err, "load toml config file")
	}
	return parse(config)
}

// parseTree parses the given toml Tree.
func parse(config *toml.Tree) error {
	if err := config.Get("App").(*toml.Tree).Unmarshal(&App); err != nil {
		return errors.Wrap(err, "mapping [App] section")
	}

	if err := config.Get("Database").(*toml.Tree).Unmarshal(&Database); err != nil {
		return errors.Wrap(err, "mapping [Database] section")
	}

	if err := config.Get("Game").(*toml.Tree).Unmarshal(&Game); err != nil {
		return errors.Wrap(err, "mapping [Game] section")
	}

	return nil
}

func Save(customConf string) error {
	if customConf == "" {
		customConf = "./conf/Cardinal.toml"
	}

	configBytes, err := toml.Marshal(map[string]interface{}{
		"App":      App,
		"Database": Database,
		"Game":     Game,
	})
	if err != nil {
		return errors.Wrap(err, "marshal")
	}

	if err := os.WriteFile(customConf, configBytes, 0644); err != nil {
		return errors.Wrap(err, "write file")
	}
	return nil
}
