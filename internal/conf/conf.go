// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package conf

import (
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	log "unknwon.dev/clog/v2"
)

func init() {
	err := log.NewConsole()
	if err != nil {
		panic("init console logger: " + err.Error())
	}
}

// File is the configuration object.
var File *ini.File

func Init(customConf string) error {
	if customConf == "" {
		customConf = "./conf/Cardinal.ini"
	}

	var err error
	File, err = ini.Load(customConf)
	if err != nil {
		return errors.Wrap(err, "load ini")
	}

	if err := File.Section("App").MapTo(&App); err != nil {
		return errors.Wrap(err, "mapping [App] section")
	}

	if err := File.Section("Database").MapTo(&Database); err != nil {
		return errors.Wrap(err, "mapping [Database] section")
	}

	if err := File.Section("Game").MapTo(&Game); err != nil {
		return errors.Wrap(err, "mapping [Game] section")
	}
	// TODO Check pause time.

	if err := File.Section("Server").MapTo(&Server); err != nil {
		return errors.Wrap(err, "mapping [Server] section")
	}

	return nil
}
