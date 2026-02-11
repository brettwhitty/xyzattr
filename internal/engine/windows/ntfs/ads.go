// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build windows

package ntfs

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procFindFirstStreamW = modkernel32.NewProc("FindFirstStreamW")
	procFindNextStreamW  = modkernel32.NewProc("FindNextStreamW")
	procFindClose        = modkernel32.NewProc("FindClose")
)

type WIN32_FIND_STREAM_DATA struct {
	StreamSize int64
	StreamName [260 + 36]uint16
}

type ADSStore struct{}

func (s *ADSStore) Get(path string, name string) ([]byte, error) {
	return os.ReadFile(fmt.Sprintf("%s:%s", path, name))
}

func (s *ADSStore) Set(path string, name string, value []byte) error {
	return os.WriteFile(fmt.Sprintf("%s:%s", path, name), value, 0666)
}

func (s *ADSStore) List(path string) ([]string, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var data WIN32_FIND_STREAM_DATA
	handle, _, err := procFindFirstStreamW.Call(uintptr(unsafe.Pointer(pathPtr)), 0, uintptr(unsafe.Pointer(&data)), 0)
	if handle == ^uintptr(0) {
		return nil, nil
	}
	defer procFindClose.Call(handle)

	var streams []string
	for {
		name := syscall.UTF16ToString(data.StreamName[:])
		if name != "" {
			parts := strings.Split(name, ":")
			if len(parts) >= 2 && parts[1] != "" {
				streams = append(streams, parts[1])
			}
		}
		res, _, _ := procFindNextStreamW.Call(handle, uintptr(unsafe.Pointer(&data)))
		if res == 0 {
			break
		}
	}
	return streams, nil
}

func (s *ADSStore) Remove(path string, name string) error {
	return os.Remove(fmt.Sprintf("%s:%s", path, name))
}
