// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build windows

package ntfs

import (
	"os"
	"path/filepath"
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

var findNextStreamW = func(handle uintptr, data *WIN32_FIND_STREAM_DATA) (uintptr, error) {
	result, _, err := procFindNextStreamW.Call(handle, uintptr(unsafe.Pointer(data)))
	return result, err
}

type ADSStore struct{}

// Resolve the directory separately so a one-character filename cannot become
// a drive qualifier, while a final component such as "file:." retains native
// stream spelling until its explicit type is appended.
func absoluteStreamBase(path string) (string, error) {
	if path == "" {
		return path, os.ErrInvalid
	}
	if filepath.IsAbs(path) {
		return path, nil
	}
	dir, base := filepath.Split(path)
	absoluteDir, err := filepath.Abs(dir)
	if err != nil {
		return path, err
	}
	return filepath.Join(absoluteDir, base), nil
}

// namedStreamPath preserves native stream syntax, including a named :$DATA
// suffix. It must never resolve to the file or directory's default stream.
func namedStreamPath(path, name string) (string, error) {
	path, err := absoluteStreamBase(path)
	if err != nil {
		return path, err
	}
	target := path + ":" + name
	// Separators can let Win32 normalize an attribute into a different path.
	if name == "" || strings.ContainsAny(name, `/\`) {
		return target, os.ErrInvalid
	}
	// Inspect the complete target: path may itself end with a stream colon.
	// Base strips a drive/UNC volume without confusing its colon with ADS.
	_, stream, found := strings.Cut(filepath.Base(target), ":")
	if !found || stream == "" || stream[0] == ':' {
		return target, os.ErrInvalid
	}
	// The directory's default stream is named $I30:$INDEX_ALLOCATION.
	// Win32 also accepts case variants and trailing dots/spaces. Reject the
	// alias without changing the target or restricting named $DATA streams.
	if strings.EqualFold(strings.TrimRight(stream, " ."), "$I30:$INDEX_ALLOCATION") {
		return target, os.ErrInvalid
	}
	return target, nil
}

func (s *ADSStore) Get(path string, name string) ([]byte, error) {
	target, err := namedStreamPath(path, name)
	if err != nil {
		return nil, &os.PathError{Op: "getxattr", Path: target, Err: err}
	}
	return os.ReadFile(target)
}

func (s *ADSStore) Set(path string, name string, value []byte) error {
	target, err := namedStreamPath(path, name)
	if err != nil {
		return &os.PathError{Op: "setxattr", Path: target, Err: err}
	}
	return os.WriteFile(target, value, 0666)
}

func (s *ADSStore) List(path string) ([]string, error) {
	path, err := absoluteStreamBase(path)
	if err != nil {
		return nil, &os.PathError{Op: "listxattr", Path: path, Err: err}
	}
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	var data WIN32_FIND_STREAM_DATA
	handle, _, err := procFindFirstStreamW.Call(uintptr(unsafe.Pointer(pathPtr)), 0, uintptr(unsafe.Pointer(&data)), 0)
	if handle == ^uintptr(0) {
		if err == syscall.ERROR_HANDLE_EOF {
			return nil, nil
		}
		return nil, &os.PathError{Op: "listxattr", Path: path, Err: err}
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
		res, err := findNextStreamW(handle, &data)
		if res == 0 {
			if err == syscall.ERROR_HANDLE_EOF {
				break
			}
			return nil, &os.PathError{Op: "listxattr", Path: path, Err: err}
		}
	}
	return streams, nil
}

func (s *ADSStore) Remove(path string, name string) error {
	target, err := namedStreamPath(path, name)
	if err != nil {
		return &os.PathError{Op: "removexattr", Path: target, Err: err}
	}
	return os.Remove(target)
}
