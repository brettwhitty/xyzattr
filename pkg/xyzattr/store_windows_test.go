// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build windows

package xyzattr

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gitea.gnomatix.com/gnomatix/xyzattr/internal/engine/windows/ntfs"
)

func nativeStoreFixture(t *testing.T) (*FileStore, func()) {
	t.Helper()
	return namedNativeStoreFixture(t, "original.txt")
}

func namedNativeStoreFixture(t *testing.T, name string) (*FileStore, func()) {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	original := []byte("FileStore original\x00\xff\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	store := At(path)
	if _, ok := store.engine.(*ntfs.ADSStore); !ok {
		t.Fatalf("NTFS fixture selected %T; test requires real native dispatch", store.engine)
	}
	check := func() {
		t.Helper()
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, original) {
			t.Errorf("original file changed or disappeared: %q, %v", got, err)
		}
	}
	t.Cleanup(check)
	return store, check
}

func TestFileStoreADSDefaultStream(t *testing.T) {
	for _, name := range []string{"", ":", ":$DATA", ":$data", ":$DaTa", ":$DATA ", ":$DATA.", "unused/../original.txt", `unused\..\original.txt`} {
		for _, op := range []string{"Get", "Set", "Remove"} {
			t.Run(fmt.Sprintf("%s/%q", op, name), func(t *testing.T) {
				store, check := nativeStoreFixture(t)
				var err error
				switch op {
				case "Get":
					var value []byte
					value, err = store.Get(name)
					if len(value) != 0 {
						t.Errorf("Get exposed file contents: %q", value)
					}
				case "Set":
					err = store.Set(name, []byte("attribute replacement"))
				case "Remove":
					err = store.Remove(name)
				}
				if err == nil {
					t.Error("FileStore accepted the default stream")
				}
				check()
			})
		}
	}
}

func TestFileStoreADSListError(t *testing.T) {
	store, check := nativeStoreFixture(t)
	names, err := At(store.path + ".missing").List()
	check()
	if !errors.Is(err, os.ErrNotExist) || len(names) != 0 {
		t.Errorf("List lost native failure: %q, %v", names, err)
	}
}

func TestFileStoreADSNativeValues(t *testing.T) {
	store, check := nativeStoreFixture(t)
	for _, value := range [][]byte{{0x00, 0xff, 0x01}, {}, []byte("replacement"), nil} {
		err := store.Set("$DATA:$DATA", value)
		check()
		if err != nil {
			t.Fatal(err)
		}
		got, err := At(store.path).Get("$DATA")
		check()
		if err != nil || !bytes.Equal(got, value) {
			t.Fatalf("Get: %v, %v; want %v", got, err, value)
		}
		names, err := store.List()
		check()
		if err != nil || !slices.Equal(names, []string{"$DATA"}) {
			t.Fatalf("List: %q, %v", names, err)
		}
	}
	err := store.Remove("$DATA:$DATA")
	check()
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Get("$DATA")
	check()
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Get removed stream: %v", err)
	}
	// A namespaced property has no defined portable mapping here. Preserve
	// the native failure instead of silently selecting a different backend.
	err = store.Set("dreamfs:property", []byte("property value"))
	check()
	if err == nil {
		t.Error("unsupported native name silently succeeded")
	}
	names, err := store.List()
	check()
	if err != nil || len(names) != 0 {
		t.Errorf("List after removal and rejected name: %q, %v", names, err)
	}
}

func TestFileStoreADSDirectoryDefaultStream(t *testing.T) {
	for _, tc := range []struct{ pathSuffix, name string }{
		{"", "$I30:$INDEX_ALLOCATION"},
		{":$I30", "$INDEX_ALLOCATION"},
	} {
		t.Run(tc.pathSuffix+tc.name, func(t *testing.T) {
			store, check := nativeStoreFixture(t)
			dir := filepath.Join(filepath.Dir(store.path), "empty")
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			err := At(dir + tc.pathSuffix).Remove(tc.name)
			check()
			if err == nil {
				t.Error("FileStore accepted the directory's default stream")
			}
			if info, err := os.Stat(dir); err != nil || !info.IsDir() {
				t.Errorf("original directory did not survive: %v", err)
			}
		})
	}
}

func TestFileStoreADSRelativeSingleCharacterFile(t *testing.T) {
	for _, form := range []string{"bare", "dot-backslash", "dot-slash", "absolute", "absolute-slash", "extended", "drive-relative", "root-relative", "nested-relative"} {
		t.Run(form, func(t *testing.T) {
			absolute, check := namedNativeStoreFixture(t, "a")
			t.Chdir(filepath.Dir(absolute.path))
			target := "a"
			switch form {
			case "dot-backslash":
				target = `.\a`
			case "dot-slash":
				target = "./a"
			case "absolute":
				target = absolute.path
			case "absolute-slash":
				target = filepath.ToSlash(absolute.path)
			case "extended":
				target = `\\?\` + absolute.path
			case "drive-relative":
				target = filepath.VolumeName(absolute.path) + "a"
			case "root-relative":
				target = absolute.path[len(filepath.VolumeName(absolute.path)):]
			case "nested-relative":
				t.Chdir(filepath.Dir(filepath.Dir(absolute.path)))
				target = filepath.Join(filepath.Base(filepath.Dir(absolute.path)), "a")
			}
			store := At(target)
			if _, ok := store.engine.(*ntfs.ADSStore); !ok {
				t.Fatalf("relative NTFS file selected %T instead of ADS", store.engine)
			}
			value := []byte("relative FileStore attribute")
			err := store.Set("user.comment", value)
			check()
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(absolute.path + ":user.comment")
			check()
			if err != nil || !bytes.Equal(got, value) {
				t.Fatalf("Set did not reach intended native stream: %q, %v", got, err)
			}
			got, err = store.Get("user.comment")
			check()
			if err != nil || !bytes.Equal(got, value) {
				t.Fatalf("Get: %q, %v", got, err)
			}
			names, err := store.List()
			check()
			if err != nil || !slices.Equal(names, []string{"user.comment"}) {
				t.Fatalf("List: %q, %v", names, err)
			}
			err = store.Remove("user.comment")
			check()
			if err != nil {
				t.Fatal(err)
			}
			_, err = os.ReadFile(absolute.path + ":user.comment")
			check()
			if !errors.Is(err, os.ErrNotExist) {
				t.Errorf("Remove left native stream: %v", err)
			}
		})
	}
}
