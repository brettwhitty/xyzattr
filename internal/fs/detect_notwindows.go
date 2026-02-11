// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !windows

package fs

func detectWindows(path string) (Type, error) {
	return Unknown, nil
}
