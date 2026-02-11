// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !linux

package fs

func detectLinux(path string) (Type, error) {
	return Unknown, nil
}
