// Copyright (c) 2026 GNOMATIX. All rights reserved.

package fallback

import (
	"os"
	"syscall"
	"unsafe"
)

var lockFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("LockFileEx")

func lockFile(f *os.File, exclusive bool) error {
	var flags uintptr
	if exclusive {
		flags = 0x00000002 // LOCKFILE_EXCLUSIVE_LOCK
	}
	var overlapped syscall.Overlapped
	// os.OpenFile opens a synchronous handle. Without FAIL_IMMEDIATELY the
	// call waits for the lock; the locked byte may extend past end of file.
	ok, _, err := lockFileEx.Call(f.Fd(), flags, 0, 1, 0, uintptr(unsafe.Pointer(&overlapped)))
	if ok == 0 {
		return err
	}
	return nil
}
