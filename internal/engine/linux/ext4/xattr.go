// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build linux

package ext4

import "syscall"

var (
	getxattr  = syscall.Getxattr
	listxattr = syscall.Listxattr
)

type XattrStore struct{}

func (s *XattrStore) Get(path string, name string) ([]byte, error) {
	size, err := getxattr(path, name, nil)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return []byte{}, nil
	}
	buf := make([]byte, size)
	n, err := getxattr(path, name, buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

func (s *XattrStore) Set(path string, name string, value []byte) error {
	return syscall.Setxattr(path, name, value, 0)
}

func (s *XattrStore) List(path string) ([]string, error) {
	size, err := listxattr(path, nil)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return []string{}, nil
	}
	buf := make([]byte, size)
	n, err := listxattr(path, buf)
	if err != nil {
		return nil, err
	}
	var names []string
	start := 0
	for i, b := range buf[:n] {
		if b == 0 {
			names = append(names, string(buf[start:i]))
			start = i + 1
		}
	}
	return names, nil
}

func (s *XattrStore) Remove(path string, name string) error {
	return syscall.Removexattr(path, name)
}
