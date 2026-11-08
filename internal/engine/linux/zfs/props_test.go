// Copyright (c) 2026 GNOMATIX. All rights reserved.

package zfs

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestUnsupportedOperations(t *testing.T) {
	store := &PropsStore{}
	operations := []struct {
		name string
		call func(*testing.T, string) error
	}{
		{"Get", func(t *testing.T, path string) error {
			value, err := store.Get(path, "user.attribute")
			if value != nil {
				t.Errorf("unsupported Get returned %x; want nil", value)
			}
			return err
		}},
		{"Set", func(t *testing.T, path string) error {
			return store.Set(path, "user.attribute", []byte("value\x00\xff"))
		}},
		{"List", func(t *testing.T, path string) error {
			names, err := store.List(path)
			if names != nil {
				t.Errorf("unsupported List returned %q; want nil", names)
			}
			return err
		}},
		{"Remove", func(t *testing.T, path string) error {
			return store.Remove(path, "user.attribute")
		}},
	}
	for _, kind := range []string{"file", "directory"} {
		t.Run(kind, func(t *testing.T) {
			for _, op := range operations {
				t.Run(op.name, func(t *testing.T) {
					root := t.TempDir()
					target := filepath.Join(root, "original")
					contentPath := target
					if kind == "directory" {
						if err := os.Mkdir(target, 0700); err != nil {
							t.Fatal(err)
						}
						contentPath = filepath.Join(target, "child")
					}
					original := []byte("original\x00content\xff")
					if err := os.WriteFile(contentPath, original, 0600); err != nil {
						t.Fatal(err)
					}
					before, err := os.Stat(target)
					if err != nil {
						t.Fatal(err)
					}
					contentBefore, err := os.Stat(contentPath)
					if err != nil {
						t.Fatal(err)
					}
					if err := op.call(t, target); !errors.Is(err, errors.ErrUnsupported) {
						t.Errorf("got %v; want errors.ErrUnsupported", err)
					}
					after, err := os.Stat(target)
					if err != nil || !os.SameFile(before, after) {
						t.Errorf("original identity changed: %v", err)
					}
					contentAfter, err := os.Stat(contentPath)
					if err != nil || !os.SameFile(contentBefore, contentAfter) {
						t.Errorf("content file identity changed: %v", err)
					}
					got, err := os.ReadFile(contentPath)
					if err != nil || !bytes.Equal(got, original) {
						t.Errorf("original content changed: %x, %v", got, err)
					}
					entries, err := os.ReadDir(root)
					if err != nil || len(entries) != 1 || entries[0].Name() != "original" {
						t.Errorf("target sidecars created: %v, %v", entries, err)
					}
					if kind == "directory" {
						children, err := os.ReadDir(target)
						if err != nil || len(children) != 1 || children[0].Name() != "child" {
							t.Errorf("directory entries changed: %v, %v", children, err)
						}
					}
				})
			}
		})
	}
}
