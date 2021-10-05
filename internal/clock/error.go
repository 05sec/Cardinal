// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package clock

import (
	"github.com/pkg/errors"
)

var (
	ErrStartTimeOrder    = errors.New("start time should before end time")
	ErrRestTimeFormat    = errors.New("rest time format error")
	ErrRestTimeOrder     = errors.New("rest start time should before end time")
	ErrRestTimeOverflow  = errors.New("rest time overflow")
	ErrRestTimeListOrder = errors.New("rest time list should in order")
)
