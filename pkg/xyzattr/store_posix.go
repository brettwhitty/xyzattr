// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !windows

package xyzattr

import (
	"gitea.gnomatix.com/gnomatix/xyzattr/internal/engine/darwin/apfs"
	"gitea.gnomatix.com/gnomatix/xyzattr/internal/engine/linux/ext4"
	"gitea.gnomatix.com/gnomatix/xyzattr/internal/fs"
)

func getNativeEngineByFS(t fs.Type) Engine {
	switch t {
	case fs.EXT4, fs.XFS, fs.BTRFS, fs.F2FS, fs.TMPFS, fs.OVERLAY, fs.JFS:
		return &ext4.XattrStore{}
	case fs.APFS, fs.HFS:
		return &apfs.XattrStore{}
	case fs.ZFS:
		// Placeholder for ZFS engine
		return nil
	default:
		return nil
	}
}
