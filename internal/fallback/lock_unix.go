// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build linux || darwin

package fallback

import (
	"os"
	"syscall"
)

func lockFile(f *os.File, exclusive bool) error {
	mode := syscall.LOCK_SH
	if exclusive {
		mode = syscall.LOCK_EX
	}
	for {
		err := syscall.Flock(int(f.Fd()), mode)
		if err != syscall.EINTR {
			return err
		}
	}
}
