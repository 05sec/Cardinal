// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package dbutil

import (
	"time"
)

func Now() time.Time {
	return time.Now().Truncate(time.Microsecond)
}
