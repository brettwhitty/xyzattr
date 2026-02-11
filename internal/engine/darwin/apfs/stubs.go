// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !darwin

package apfs

import "errors"

type XattrStore struct{}

func (s *XattrStore) Get(path string, name string) ([]byte, error) {
	return nil, errors.New("not implemented")
}
func (s *XattrStore) Set(path string, name string, value []byte) error {
	return errors.New("not implemented")
}
func (s *XattrStore) List(path string) ([]string, error)    { return nil, nil }
func (s *XattrStore) Remove(path string, name string) error { return nil }
