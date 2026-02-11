// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build windows

package fs

import (
	"syscall"
	"unsafe"
)

func detectWindows(path string) (Type, error) {
	volumePath := path
	if len(path) >= 3 && path[1] == ':' && path[2] == '\\' {
		volumePath = path[:3]
	}

	pathPtr, err := syscall.UTF16PtrFromString(volumePath)
	if err != nil {
		return Unknown, err
	}

	var fsName [260]uint16
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetVolumeInformationW := modkernel32.NewProc("GetVolumeInformationW")

	ret, _, err := procGetVolumeInformationW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
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
