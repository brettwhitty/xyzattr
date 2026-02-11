// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build linux

package fs

import (
	"syscall"
)

func detectLinux(path string) (Type, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return Unknown, err
	}

	switch stat.Type {
	case 0xef53: // EXT4_SUPER_MAGIC
		return EXT4, nil
	case 0x58465342: // XFS_SUPER_MAGIC
		return XFS, nil
	case 0x9123683e: // BTRFS_SUPER_MAGIC
		return BTRFS, nil
	case 0xf2f52010: // F2FS_SUPER_MAGIC
		return F2FS, nil
	case 0x01021994: // TMPFS_MAGIC
		return TMPFS, nil
	case 0x794c7630: // OVERLAYFS_MAGIC
		return OVERLAY, nil
	case 0x3153464a: // JFS_SUPER_MAGIC
		return JFS, nil
	case 0x2fc12fc1: // ZFS_SUPER_MAGIC
		return ZFS, nil
	default:
		return Unknown, nil
	}
}
