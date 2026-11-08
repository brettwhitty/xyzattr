// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build !linux

package ext4

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestUnsupportedOperations(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "original")
	original := []byte("original\x00content\xff")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	s := &XattrStore{}
	for _, op := range []struct {
		name string
		call func() error
	}{
		{"Get", func() error {
			value, e := s.Get(path, "attribute")
			if value != nil {
				t.Errorf("unsupported Get returned %x", value)
			}
			return e
		}},
		{"Set", func() error { return s.Set(path, "attribute", []byte("value")) }},
		{"List", func() error {
			names, e := s.List(path)
			if names != nil {
				t.Errorf("unsupported List returned %q", names)
			}
			return e
		}},
		{"Remove", func() error { return s.Remove(path, "attribute") }},
	} {
		t.Run(op.name, func(t *testing.T) {
			if err := op.call(); !errors.Is(err, errors.ErrUnsupported) {
				t.Errorf("got %v; want ErrUnsupported", err)
			}
		})
	}
	current, err := os.Stat(path)
	if err != nil || !os.SameFile(info, current) {
		t.Errorf("original identity changed: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, original) {
		t.Errorf("original content changed: %x, %v", got, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != "original" {
		t.Errorf("target sidecars created: %v, %v", entries, err)
	}
}
