// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build darwin

package fs

import (
	"syscall"
)

func detectDarwin(path string) (Type, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return Unknown, err
	}

	name := ""
	for _, b := range stat.Fstypename {
		if b == 0 {
			break
		}
		name += string(b)
	}

	switch name {
	case "apfs":
		return APFS, nil
	case "hfs":
		return HFS, nil
	default:
		return Type(name), nil
	}
}
