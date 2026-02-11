// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build linux

package ext4

import (
	"syscall"
	"unsafe"
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
	_, err = getxattr(path, name, &buf[0])
	return buf, err
}

func (s *XattrStore) Set(path string, name string, value []byte) error {
	var ptr *byte
	if len(value) > 0 {
		ptr = &value[0]
	}
	return setxattr(path, name, ptr, len(value))
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
	_, err = listxattr(path, &buf[0])
	if err != nil {
		return nil, err
	}
	var names []string
	start := 0
	for i, b := range buf {
		if b == 0 {
			names = append(names, string(buf[start:i]))
			start = i + 1
		}
	}
	return names, nil
}

func (s *XattrStore) Remove(path string, name string) error {
	return removexattr(path, name)
}

// Syscall wrappers (Linux specific as per spec hierarchy)
func getxattr(path string, name string, ptr *byte) (int, error) {
	pPtr, _ := syscall.BytePtrFromString(path)
	nPtr, _ := syscall.BytePtrFromString(name)
	size, _, errno := syscall.Syscall6(syscall.SYS_GETXATTR, uintptr(unsafe.Pointer(pPtr)), uintptr(unsafe.Pointer(nPtr)), uintptr(unsafe.Pointer(ptr)), 0, 0, 0)
	if errno != 0 {
		return 0, errno
	}
	return int(size), nil
}

func setxattr(path string, name string, ptr *byte, size int) error {
	pPtr, _ := syscall.BytePtrFromString(path)
	nPtr, _ := syscall.BytePtrFromString(name)
	_, _, errno := syscall.Syscall6(syscall.SYS_SETXATTR, uintptr(unsafe.Pointer(pPtr)), uintptr(unsafe.Pointer(nPtr)), uintptr(unsafe.Pointer(ptr)), uintptr(size), 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}

func listxattr(path string, ptr *byte) (int, error) {
	pPtr, _ := syscall.BytePtrFromString(path)
	size, _, errno := syscall.Syscall(syscall.SYS_LISTXATTR, uintptr(unsafe.Pointer(pPtr)), uintptr(unsafe.Pointer(ptr)), 0)
	if errno != 0 {
		return 0, errno
	}
	return int(size), nil
}

func removexattr(path string, name string) error {
	pPtr, _ := syscall.BytePtrFromString(path)
	nPtr, _ := syscall.BytePtrFromString(name)
	_, _, errno := syscall.Syscall(syscall.SYS_REMOVEXATTR, uintptr(unsafe.Pointer(pPtr)), uintptr(unsafe.Pointer(nPtr)), 0)
	if errno != 0 {
		return errno
	}
	return nil
}
