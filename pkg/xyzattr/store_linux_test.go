// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build linux

package xyzattr

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"

	"gitea.gnomatix.com/gnomatix/xyzattr/internal/engine/linux/ext4"
	"gitea.gnomatix.com/gnomatix/xyzattr/internal/fs"
)

func TestFileStoreLinuxNativeDispatch(t *testing.T) {
	fixtureHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	registryPath := filepath.Join(fixtureHome, ".gnomatix", "xyzattr", "attributes.json")
	registryBefore, registryErr := os.ReadFile(registryPath)
	if registryErr != nil && !errors.Is(registryErr, os.ErrNotExist) {
		t.Fatal(registryErr)
	}
	t.Cleanup(func() {
		registryAfter, err := os.ReadFile(registryPath)
		if registryErr == nil {
			if err != nil || !bytes.Equal(registryAfter, registryBefore) {
				t.Errorf("native calls changed existing fallback database: %v", err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("native calls created fallback database or read failed: %v", err)
		}
	})
	path := filepath.Join(t.TempDir(), "original")
	original := []byte("original\x00contents\xff")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	check := func() {
		t.Helper()
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, original) {
			t.Errorf("original content: %x, %v", got, err)
		}
		current, err := os.Stat(path)
		if err != nil || !os.SameFile(info, current) {
			t.Errorf("original identity changed: %v", err)
		}
	}
	t.Cleanup(check)
	kind, err := fs.Detect(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("fixture=%s detected_filesystem=%s", path, kind)
	store := At(path)
	if _, ok := store.engine.(*ext4.XattrStore); !ok {
		t.Fatalf("native fixture selected %T, filesystem=%s", store.engine, kind)
	}
	wantNames := []string{"user.binary", "user.empty", "user.namespace:property"}
	values := [][]byte{{0, 0xff, 2, 0}, {}, []byte("native")}
	for i, name := range wantNames {
		if err := store.Set(name, values[i]); err != nil {
			t.Fatal(err)
		}
		check()
		buf := make([]byte, 128)
		n, err := syscall.Getxattr(path, name, buf)
		if err != nil || !bytes.Equal(buf[:max(n, 0)], values[i]) {
			t.Fatalf("native read %s: %x, %v", name, buf, err)
		}
		got, err := At(path).Get(name)
		check()
		if err != nil || !bytes.Equal(got, values[i]) {
			t.Errorf("FileStore.Get %s: %x, %v; want %x", name, got, err, values[i])
		}
	}
	names, err := store.List()
	check()
	slices.Sort(names)
	if err != nil || !slices.Equal(names, wantNames) {
		t.Errorf("FileStore.List: %q, %v; want %q", names, err, wantNames)
	}
	if _, err := store.Get("user.absent"); !errors.Is(err, syscall.ENODATA) {
		t.Errorf("missing Get: %v", err)
	}
	check()
	if err := store.Remove("user.absent"); !errors.Is(err, syscall.ENODATA) {
		t.Errorf("missing Remove: %v", err)
	}
	check()
	for _, op := range []struct {
		name string
		call func() error
	}{
		{"Get", func() error { _, err := store.Get("user.binary\x00suffix"); return err }},
		{"Set", func() error { return store.Set("user.binary\x00suffix", []byte("bad")) }},
		{"Remove", func() error { return store.Remove("user.binary\x00suffix") }},
	} {
		if err := op.call(); !errors.Is(err, syscall.EINVAL) {
			t.Errorf("NUL %s: %v; want EINVAL", op.name, err)
		}
		check()
	}
	for _, name := range wantNames {
		if err := store.Remove(name); err != nil {
			t.Fatal(err)
		}
		check()
		if _, err := syscall.Getxattr(path, name, nil); !errors.Is(err, syscall.ENODATA) {
			t.Errorf("native read after removal %s: %v", name, err)
		}
	}
	if names, err := store.List(); err != nil || len(names) != 0 {
		t.Errorf("empty FileStore.List: %q, %v", names, err)
	}
	check()
}
