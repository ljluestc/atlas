// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package migrate

import (
	"errors"
)

// ErrNonLinearHistory is returned when migrations are non-linear.
var ErrNonLinearHistory = errors.New("sql/migrate: migration history is non-linear")
