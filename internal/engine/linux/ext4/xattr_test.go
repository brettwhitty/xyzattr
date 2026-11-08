// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build linux

package ext4

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
)

func xattrFixture(t *testing.T) (string, func()) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "original")
	original := []byte{'o', 'r', 'i', 'g', 0, 0xff, '\n'}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		t.Fatal(err)
	}
	t.Logf("fixture=%s filesystem_magic=%#x", path, stat.Type)
	check := func() {
		t.Helper()
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, original) {
			t.Errorf("original contents: got %x, err %v; want %x", got, err, original)
		}
		current, err := os.Stat(path)
		if err != nil || !os.SameFile(info, current) {
			t.Errorf("original file identity changed: %v", err)
		}
	}
	t.Cleanup(check)
	return path, check
}

func TestXattrNativeRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value []byte
	}{
		{"user.binary", []byte{0, 0xff, 1, 0, 0x80}},
		{"user.empty", []byte{}},
		{"user.nil", nil},
		{"user.namespace:property", []byte("colon")},
		{"user.space slash/日本", []byte("native name")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path, check := xattrFixture(t)
			store := &XattrStore{}
			if err := store.Set(path, tc.name, tc.value); err != nil {
				t.Fatal(err)
			}
			check()
			native := make([]byte, 128)
			n, err := syscall.Getxattr(path, tc.name, native)
			if err != nil || !bytes.Equal(native[:max(n, 0)], tc.value) {
				t.Fatalf("native read: got %x, err %v; want %x", native[:max(n, 0)], err, tc.value)
			}
			got, err := store.Get(path, tc.name)
			check()
			if err != nil || !bytes.Equal(got, tc.value) {
				t.Errorf("Get: got %x, err %v; want %x", got, err, tc.value)
			}
			if err := syscall.Setxattr(path, tc.name, []byte("replaced"), 0); err != nil {
				t.Fatal(err)
			}
			got, err = store.Get(path, tc.name)
			check()
			if err != nil || string(got) != "replaced" {
				t.Errorf("Get native replacement: got %q, err %v", got, err)
			}
			if err := store.Set(path, tc.name, nil); err != nil {
				t.Fatal(err)
			}
			check()
			got, err = store.Get(path, tc.name)
			check()
			if err != nil || len(got) != 0 {
				t.Errorf("Get after replacement with empty value: %x, %v", got, err)
			}
			if n, err := syscall.Getxattr(path, tc.name, nil); err != nil || n != 0 {
				t.Errorf("native empty replacement: size %d, err %v", n, err)
			}
			if err := store.Remove(path, tc.name); err != nil {
				t.Fatal(err)
			}
			check()
			if _, err := syscall.Getxattr(path, tc.name, nil); !errors.Is(err, syscall.ENODATA) {
				t.Errorf("native read after removal: %v", err)
			}
			if got, err := store.Get(path, tc.name); got != nil || !errors.Is(err, syscall.ENODATA) {
				t.Errorf("missing Get: got %x, err %v", got, err)
			}
			check()
			if err := store.Remove(path, tc.name); !errors.Is(err, syscall.ENODATA) {
				t.Errorf("missing Remove: %v", err)
			}
			check()
		})
	}
}

func TestXattrListNativeNames(t *testing.T) {
	path, check := xattrFixture(t)
	store := &XattrStore{}
	if got, err := store.List(path); err != nil || len(got) != 0 {
		t.Fatalf("empty List: %q, %v", got, err)
	}
	check()
	want := []string{"user.binary", "user.empty", "user.namespace:property", "user.space slash/日本"}
	for _, name := range want {
		if err := syscall.Setxattr(path, name, nil, 0); err != nil {
			t.Fatal(err)
		}
		check()
	}
	got, err := store.List(path)
	check()
	slices.Sort(got)
	slices.Sort(want)
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("List: got %q, err %v; want %q", got, err, want)
	}
}

func TestXattrDirectoryRoundTrip(t *testing.T) {
	child, checkChild := xattrFixture(t)
	path := filepath.Dir(child)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	check := func() {
		t.Helper()
		checkChild()
		current, err := os.Stat(path)
		if err != nil || !current.IsDir() || !os.SameFile(info, current) {
			t.Errorf("original directory changed: %v", err)
		}
		entries, err := os.ReadDir(path)
		if err != nil || len(entries) != 1 || entries[0].Name() != "original" {
			t.Errorf("original directory entries changed: %v, %v", entries, err)
		}
	}
	t.Cleanup(check)
	store := &XattrStore{}
	wantNames := []string{"user.directory", "user.empty"}
	values := [][]byte{{0xff, 0, 1}, {}}
	for i, name := range wantNames {
		if err := store.Set(path, name, values[i]); err != nil {
			t.Fatal(err)
		}
		check()
		got, err := store.Get(path, name)
		check()
		if err != nil || !bytes.Equal(got, values[i]) {
			t.Errorf("directory Get %s: %x, %v; want %x", name, got, err, values[i])
		}
		buf := make([]byte, 16)
		n, err := syscall.Getxattr(path, name, buf)
		if err != nil || !bytes.Equal(buf[:max(n, 0)], values[i]) {
			t.Errorf("native directory value %s: %x, %v", name, buf, err)
		}
	}
	names, err := store.List(path)
	check()
	slices.Sort(names)
	if err != nil || !slices.Equal(names, wantNames) {
		t.Errorf("directory List: %q, %v; want %q", names, err, wantNames)
	}
	for _, name := range wantNames {
		if err := store.Remove(path, name); err != nil {
			t.Fatal(err)
		}
		check()
		if _, err := syscall.Getxattr(path, name, nil); !errors.Is(err, syscall.ENODATA) {
			t.Errorf("native directory removal %s: %v", name, err)
		}
		if got, err := store.Get(path, name); got != nil || !errors.Is(err, syscall.ENODATA) {
			t.Errorf("missing directory Get %s: %x, %v", name, got, err)
		}
		check()
		if err := store.Remove(path, name); !errors.Is(err, syscall.ENODATA) {
			t.Errorf("missing directory Remove %s: %v", name, err)
		}
		check()
	}
	if names, err := store.List(path); err != nil || len(names) != 0 {
		t.Errorf("empty directory List: %q, %v", names, err)
	}
	check()
}

func TestXattrInvalidStringsPreserveNativeErrors(t *testing.T) {
	path, check := xattrFixture(t)
	store := &XattrStore{}
	if err := syscall.Setxattr(path, "user.kept", []byte("kept"), 0); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		label, path, name string
		want              error
	}{
		{"NUL path", path + "\x00suffix", "user.kept", syscall.EINVAL},
		{"NUL name", path, "user.kept\x00suffix", syscall.EINVAL},
		{"missing path", path + "-missing", "user.kept", syscall.ENOENT},
	} {
		t.Run(tc.label, func(t *testing.T) {
			got, err := store.Get(tc.path, tc.name)
			check()
			if got != nil || !errors.Is(err, tc.want) {
				t.Errorf("Get: %x, %v; want %v", got, err, tc.want)
			}
			if err := store.Set(tc.path, tc.name, []byte("bad")); !errors.Is(err, tc.want) {
				t.Errorf("Set: %v; want %v", err, tc.want)
			}
			check()
			if err := store.Remove(tc.path, tc.name); !errors.Is(err, tc.want) {
				t.Errorf("Remove: %v; want %v", err, tc.want)
			}
			check()
			if tc.path != path {
				if got, err := store.List(tc.path); got != nil || !errors.Is(err, tc.want) {
					t.Errorf("List: %q, %v; want %v", got, err, tc.want)
				}
				check()
			}
			buf := make([]byte, 16)
			n, err := syscall.Getxattr(path, "user.kept", buf)
			if err != nil || string(buf[:max(n, 0)]) != "kept" {
				t.Fatalf("valid attribute changed: %q, %v", buf, err)
			}
		})
	}
}

func TestXattrGetChangesBetweenSizeAndRead(t *testing.T) {
	for _, tc := range []struct {
		label   string
		value   []byte
		remove  bool
		want    []byte
		wantErr error
	}{
		{label: "shrink", value: []byte{0xff, 0}, want: []byte{0xff, 0}},
		{label: "shrink to empty", value: []byte{}, want: []byte{}},
		{label: "growth", value: []byte("longer than initial"), wantErr: syscall.ERANGE},
		{label: "remove", remove: true, wantErr: syscall.ENODATA},
	} {
		t.Run(tc.label, func(t *testing.T) {
			path, check := xattrFixture(t)
			if err := syscall.Setxattr(path, "user.changing", []byte("initial"), 0); err != nil {
				t.Fatal(err)
			}
			previous := getxattr
			t.Cleanup(func() { getxattr = previous })
			calls := 0
			getxattr = func(path, name string, buf []byte) (int, error) {
				calls++
				n, err := syscall.Getxattr(path, name, buf)
				if calls == 1 && err == nil {
					if len(buf) != 0 {
						t.Fatal("first call must query the native size")
					}
					if tc.remove {
						err = syscall.Removexattr(path, name)
					} else {
						err = syscall.Setxattr(path, name, tc.value, 0)
					}
					if err != nil {
						t.Fatalf("native mutation: %v", err)
					}
					check()
				}
				return n, err
			}
			got, err := (&XattrStore{}).Get(path, "user.changing")
			check()
			if !errors.Is(err, tc.wantErr) || !bytes.Equal(got, tc.want) {
				t.Errorf("Get after %s: %x, %v; want %x, %v", tc.label, got, err, tc.want, tc.wantErr)
			}
			if err != nil && got != nil {
				t.Errorf("failed native read returned unfilled buffer %x", got)
			}
			if calls != 2 {
				t.Errorf("size/read calls = %d; want 2 without retry", calls)
			}
		})
	}
}

func TestXattrListChangesBetweenSizeAndRead(t *testing.T) {
	for _, tc := range []struct {
		label   string
		remove  []string
		add     string
		want    []string
		wantErr error
	}{
		{label: "shrink", remove: []string{"user.second"}, want: []string{"user.first"}},
		{label: "shrink to empty", remove: []string{"user.first", "user.second"}, want: []string{}},
		{label: "growth", add: "user.third", wantErr: syscall.ERANGE},
	} {
		t.Run(tc.label, func(t *testing.T) {
			path, check := xattrFixture(t)
			for _, name := range []string{"user.first", "user.second"} {
				if err := syscall.Setxattr(path, name, nil, 0); err != nil {
					t.Fatal(err)
				}
			}
			previous := listxattr
			t.Cleanup(func() { listxattr = previous })
			calls := 0
			listxattr = func(path string, buf []byte) (int, error) {
				calls++
				n, err := syscall.Listxattr(path, buf)
				if calls == 1 && err == nil {
					if len(buf) != 0 {
						t.Fatal("first call must query the native list size")
					}
					for _, name := range tc.remove {
						if err := syscall.Removexattr(path, name); err != nil {
							t.Fatal(err)
						}
					}
					if tc.add != "" {
						if err := syscall.Setxattr(path, tc.add, nil, 0); err != nil {
							t.Fatal(err)
						}
					}
					check()
				}
				return n, err
			}
			got, err := (&XattrStore{}).List(path)
			check()
			slices.Sort(got)
			if !errors.Is(err, tc.wantErr) || !slices.Equal(got, tc.want) {
				t.Errorf("List after %s: %q, %v; want %q, %v", tc.label, got, err, tc.want, tc.wantErr)
			}
			if err != nil && got != nil {
				t.Errorf("failed native list returned names %q", got)
			}
			if calls != 2 {
				t.Errorf("size/read calls = %d; want 2 without retry", calls)
			}
		})
	}
}
