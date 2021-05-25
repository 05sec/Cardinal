// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package general

import (
	"github.com/vidar-team/Cardinal/internal/context"
)

func Hello(ctx context.Context) error {
	return ctx.Success("Hello Cardinal!")
}
