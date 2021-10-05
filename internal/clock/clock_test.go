// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_checkConfig(t *testing.T) {
	for _, tc := range []struct {
		name  string
		clock *Clock
		error error
	}{
		{
			name: "normal",
			clock: &Clock{
				StartAt: date(2021, 10, 3, 12, 0, 0),
				EndAt:   date(2021, 10, 5, 12, 0, 0),
				RestTime: [][]time.Time{
					{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 4, 8, 0, 0)},
					{date(2021, 10, 4, 20, 0, 0), date(2021, 10, 5, 8, 0, 0)},
				},
			},
			error: nil,
		},
		{
			name: "same start and end time",
			clock: &Clock{
				StartAt: date(2021, 10, 3, 12, 0, 0),
				EndAt:   date(2021, 10, 3, 12, 0, 0),
			},
			error: nil,
		},
		{
			name: "start time order",
			clock: &Clock{
				StartAt: date(2021, 10, 3, 12, 0, 0),
				EndAt:   date(2021, 10, 3, 5, 0, 0),
			},
			error: ErrStartTimeOrder,
		},
		{
			name: "rest time format",
			clock: &Clock{
				StartAt: date(2021, 10, 3, 12, 0, 0),
				EndAt:   date(2021, 10, 5, 12, 0, 0),
				RestTime: [][]time.Time{
					{date(2021, 10, 3, 20, 0, 0)},
				},
			},
			error: ErrRestTimeFormat,
		},
		{
			name: "rest time format",
			clock: &Clock{
				StartAt: date(2021, 10, 3, 12, 0, 0),
				EndAt:   date(2021, 10, 5, 12, 0, 0),
				RestTime: [][]time.Time{
					{date(2021, 10, 2, 20, 0, 0), date(2021, 10, 4, 8, 0, 0)},
					{date(2021, 10, 4, 20, 0, 0), date(2021, 10, 6, 8, 0, 0)},
				},
			},
			error: ErrRestTimeOverflow,
		},
		{
			name: "rest time order",
			clock: &Clock{
				StartAt: date(2021, 10, 3, 12, 0, 0),
				EndAt:   date(2021, 10, 5, 12, 0, 0),
				RestTime: [][]time.Time{
					{date(2021, 10, 4, 20, 0, 0), date(2021, 10, 5, 8, 0, 0)},
					{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 4, 8, 0, 0)},
				},
			},
			error: ErrRestTimeListOrder,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.clock.checkConfig()
			assert.Equal(t, tc.error, got)
		})
	}
}

func Test_combineDuration(t *testing.T) {
	for _, tc := range []struct {
		name      string
		durations [][]time.Time
		want      [][]time.Time
	}{
		{
			name:      "empty duration time",
			durations: [][]time.Time{},
			want:      [][]time.Time{},
		},
		{
			name: "no overlap",
			durations: [][]time.Time{
				{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 4, 8, 0, 0)},
				{date(2021, 10, 4, 20, 0, 0), date(2021, 10, 5, 8, 0, 0)},
			},
			want: [][]time.Time{
				{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 4, 8, 0, 0)},
				{date(2021, 10, 4, 20, 0, 0), date(2021, 10, 5, 8, 0, 0)},
			},
		},
		{
			name: "former includes latter",
			durations: [][]time.Time{
				{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 5, 20, 0, 0)},
				{date(2021, 10, 4, 8, 0, 0), date(2021, 10, 5, 8, 0, 0)},
			},
			want: [][]time.Time{
				{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 5, 20, 0, 0)},
			},
		},
		{
			name: "overlap",
			durations: [][]time.Time{
				{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 4, 20, 0, 0)},
				{date(2021, 10, 4, 8, 0, 0), date(2021, 10, 5, 8, 0, 0)},
			},
			want: [][]time.Time{
				{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 5, 8, 0, 0)},
			},
		},
		{
			name: "complex case",
			durations: [][]time.Time{
				{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 4, 8, 0, 0)},
				{date(2021, 10, 4, 0, 0, 0), date(2021, 10, 4, 8, 0, 0)},
				{date(2021, 10, 4, 13, 0, 0), date(2021, 10, 5, 8, 0, 0)},
				{date(2021, 10, 4, 15, 0, 0), date(2021, 10, 5, 0, 0, 0)},
			},
			want: [][]time.Time{
				{date(2021, 10, 3, 20, 0, 0), date(2021, 10, 4, 8, 0, 0)},
				{date(2021, 10, 4, 13, 0, 0), date(2021, 10, 5, 8, 0, 0)},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := combineDuration(tc.durations)
			assert.Equal(t, tc.want, got)
		})
	}
}

func date(year, month, day, hour, min, sec int) time.Time {
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
}
