// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !linux && !freebsd && !solaris

package zfs

type PropsStore struct{}

func (s *PropsStore) Get(path string, name string) ([]byte, error)     { return nil, nil }
func (s *PropsStore) Set(path string, name string, value []byte) error { return nil }
func (s *PropsStore) List(path string) ([]string, error)               { return nil, nil }
func (s *PropsStore) Remove(path string, name string) error            { return nil }
