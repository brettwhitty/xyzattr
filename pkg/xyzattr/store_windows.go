// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build windows

package xyzattr

import (
	"gitea.gnomatix.com/gnomatix/xyzattr/internal/engine/windows/ntfs"
	"gitea.gnomatix.com/gnomatix/xyzattr/internal/fs"
)

func getNativeEngineByFS(t fs.Type) Engine {
	switch t {
	case fs.NTFS:
		return &ntfs.ADSStore{}
	default:
		// Other Windows filesystems like FAT32/exFAT don't support ADS,
		// and will fall back to the registry store in the CompositeStore.
		return nil
	}
}
