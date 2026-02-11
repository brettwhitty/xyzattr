// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build linux || darwin

package fs

import (
	"syscall"
)

func detectPOSIX(path string) (Type, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return Unknown, err
	}

	// Magic numbers vary by OS and FS
	// This is a simplified mapping
	switch stat.Type {
	case 0xef53: // EXT4_SUPER_MAGIC
		return EXT4, nil
	case 0x58465342: // XFS_SUPER_MAGIC
		return "xfs", nil
	case 0x2fc12fc1: // ZFS_SUPER_MAGIC
		return ZFS, nil
	case 16405: // HFS+ on some systems
		return HFS, nil
	case 1836311910: // APFS on Darwin
		return APFS, nil
	default:
		return Unknown, nil
	}
}
