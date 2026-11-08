// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !linux && !freebsd && !solaris

package zfs

import "errors"

type PropsStore struct{}

func (s *PropsStore) Get(path string, name string) ([]byte, error) {
	return nil, errors.ErrUnsupported
}
func (s *PropsStore) Set(path string, name string, value []byte) error {
	return errors.ErrUnsupported
}
func (s *PropsStore) List(path string) ([]string, error)    { return nil, errors.ErrUnsupported }
func (s *PropsStore) Remove(path string, name string) error { return errors.ErrUnsupported }
