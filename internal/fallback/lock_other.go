// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !windows && !linux && !darwin

package fallback

import (
	"errors"
	"os"
)

func lockFile(_ *os.File, _ bool) error {
	return errors.ErrUnsupported
}
