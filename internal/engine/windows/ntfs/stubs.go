// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !windows

package ntfs

import "errors"

type WIN32_FIND_STREAM_DATA struct{}

type ADSStore struct{}

func (s *ADSStore) Get(path string, name string) ([]byte, error) {
	return nil, errors.ErrUnsupported
}
func (s *ADSStore) Set(path string, name string, value []byte) error {
	return errors.ErrUnsupported
}
func (s *ADSStore) List(path string) ([]string, error)    { return nil, errors.ErrUnsupported }
func (s *ADSStore) Remove(path string, name string) error { return errors.ErrUnsupported }
