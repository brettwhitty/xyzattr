// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !darwin

package fs

func detectDarwin(path string) (Type, error) {
	return Unknown, nil
}
