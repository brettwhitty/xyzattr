// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build windows

package fs

import (
	"path/filepath"
	"syscall"
	"unsafe"
)

func detectWindows(path string) (Type, error) {
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return Unknown, err
	}

	pathUTF16, err := syscall.UTF16FromString(fullPath)
	if err != nil {
		return Unknown, err
	}

	var fsName [260]uint16
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetVolumePathNameW := modkernel32.NewProc("GetVolumePathNameW")
	procGetVolumeInformationW := modkernel32.NewProc("GetVolumeInformationW")

	// A mount point is a prefix of the absolute path, plus a trailing slash.
	// Resolve it natively so relative paths and mounted volumes select the
	// filesystem of the target rather than a guessed drive root.
	volumePath := make([]uint16, len(pathUTF16)+1)
	ret, _, err := procGetVolumePathNameW.Call(
		uintptr(unsafe.Pointer(&pathUTF16[0])),
		uintptr(unsafe.Pointer(&volumePath[0])),
		uintptr(len(volumePath)),
	)
	if ret == 0 {
		return Unknown, err
	}
	ret, _, err = procGetVolumeInformationW.Call(
		uintptr(unsafe.Pointer(&volumePath[0])),
		0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(&fsName[0])),
		uintptr(len(fsName)),
	)

	if ret == 0 {
		return Unknown, err
	}

	name := syscall.UTF16ToString(fsName[:])
	switch name {
	case "NTFS":
		return NTFS, nil
	default:
		return Type(name), nil
	}
}
