// Copyright (c) 2026 GNOMATIX. All rights reserved.

package fs

import (
	"runtime"
)

type Type string

const (
	NTFS    Type = "ntfs"
	EXT4    Type = "ext4"
	XFS     Type = "xfs"
	BTRFS   Type = "btrfs"
	ZFS     Type = "zfs"
	F2FS    Type = "f2fs"
	TMPFS   Type = "tmpfs"
	OVERLAY Type = "overlay"
	JFS     Type = "jfs"
	APFS    Type = "apfs"
	HFS     Type = "hfs"
	Unknown Type = "unknown"
)

// Detect returns the filesystem type for the given path.
// Implementation is split across OS-specific files.
func Detect(path string) (Type, error) {
	switch runtime.GOOS {
	case "windows":
		return detectWindows(path)
	case "linux":
		return detectLinux(path)
	case "darwin":
		return detectDarwin(path)
	default:
		return Unknown, nil
	}
}
