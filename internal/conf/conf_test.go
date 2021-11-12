// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package conf

import (
	"testing"

	"github.com/Cardinal-Platform/testify/assert"
)

func TestNewInit(t *testing.T) {
	assert.Nil(t, Init("./testdata/custom.toml"))
}
