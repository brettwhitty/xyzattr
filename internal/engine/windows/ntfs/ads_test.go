// Copyright (c) 2026 GNOMATIX. All rights reserved.

//go:build windows

package ntfs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
)

// A fresh file per test keeps even a failing destructive regression isolated.
func adsFixture(t *testing.T) (string, func()) {
	t.Helper()
	return adsNamedFixture(t, "original.txt")
}

func adsNamedFixture(t *testing.T, name string) (string, func()) {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	original := []byte("original file contents\x00\xff\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	check := func() {
		t.Helper()
		got, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("original file did not survive: %v", err)
		} else if !bytes.Equal(got, original) {
			t.Errorf("original content changed: got %q, want %q", got, original)
		}
	}
	t.Cleanup(check)
	return path, check
}

func TestADSRejectsDefaultStream(t *testing.T) {
	store := &ADSStore{}
	for _, name := range []string{"", ":", ":$DATA", ":$data", ":$DaTa", ":$DATA ", ":$DATA.", "unused/../original.txt", `unused\..\original.txt`} {
		for _, op := range []string{"Get", "Set", "Remove"} {
			t.Run(fmt.Sprintf("%s/%q", op, name), func(t *testing.T) {
				path, check := adsFixture(t)
				var err error
				switch op {
				case "Get":
					var value []byte
					value, err = store.Get(path, name)
					if len(value) != 0 {
						t.Errorf("Get exposed default-stream content: %q", value)
					}
				case "Set":
					err = store.Set(path, name, []byte("attribute replacement"))
				case "Remove":
					err = store.Remove(path, name)
				}
				if err == nil {
					t.Error("default stream accepted as an attribute")
				}
				check()
			})
		}
	}
}

func TestADSRejectsDefaultStreamInBasePath(t *testing.T) {
	store := &ADSStore{}
	for _, prefix := range []string{"", `\\?\`} {
		for _, op := range []string{"Get", "Set", "Remove"} {
			t.Run(fmt.Sprintf("%q/%s", prefix, op), func(t *testing.T) {
				path, check := adsFixture(t)
				target := prefix + path + ":"
				var err error
				switch op {
				case "Get":
					var value []byte
					value, err = store.Get(target, "$DATA")
					if len(value) != 0 {
						t.Errorf("Get exposed file contents: %q", value)
					}
				case "Set":
					err = store.Set(target, "$DATA", []byte("replacement"))
				case "Remove":
					err = store.Remove(target, "$DATA")
				}
				check()
				if err == nil {
					t.Error("default stream in base path accepted")
				}
			})
		}
	}
}

func TestADSNativeRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name, listed string
		extendedPath bool
	}{
		{"user.comment", "user.comment", false},
		{"$DATA", "$DATA", false},
		{"$DATA:$DATA", "$DATA", false},
		{"$I30", "$I30", false},
		{"$I30:$DATA", "$I30", false},
		{"user.comment:$DATA", "user.comment", false},
		{"user.comment:$data", "user.comment", false},
		{"spaced name", "spaced name", false},
		{"méta.日本語", "méta.日本語", false},
		{" spaced name ", " spaced name ", true},
		{"user.comment.", "user.comment.", true},
		{".", ".", true},
		{"..", "..", true},
		{" ", " ", true},
		{". ", ". ", true},
		{"...", "...", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path, check := adsFixture(t)
			if tc.extendedPath {
				path = `\\?\` + path
			}
			store := &ADSStore{}
			for _, value := range [][]byte{{0x00, 0xff, 0x01, 0x7f}, {}, []byte("replacement"), nil} {
				err := store.Set(path, tc.name, value)
				check()
				if err != nil {
					t.Fatalf("Set: %v", err)
				}
				// Reopen via a new engine and the name exposed by enumeration.
				got, err := (&ADSStore{}).Get(path, tc.listed)
				check()
				if err != nil || !bytes.Equal(got, value) {
					t.Fatalf("Get: got %v, %v; want %v", got, err, value)
				}
				names, err := store.List(path)
				check()
				if err != nil || !slices.Equal(names, []string{tc.listed}) {
					t.Fatalf("List: got %q, %v; want [%q]", names, err, tc.listed)
				}
			}
			err := store.Remove(path, tc.name)
			check()
			if err != nil {
				t.Fatal(err)
			}
			_, err = store.Get(path, tc.listed)
			check()
			if !errors.Is(err, os.ErrNotExist) {
				t.Errorf("Get removed stream: got %v, want not-exist", err)
			}
			err = store.Remove(path, tc.name)
			check()
			if !errors.Is(err, os.ErrNotExist) {
				t.Errorf("Remove missing stream: got %v, want not-exist", err)
			}
			names, err := store.List(path)
			check()
			if err != nil || len(names) != 0 {
				t.Errorf("List after removal: %q, %v", names, err)
			}
		})
	}
}

func TestADSNativeErrors(t *testing.T) {
	for _, name := range []string{"dreamfs:property", "invalid\x00name"} {
		t.Run(fmt.Sprintf("%q", name), func(t *testing.T) {
			path, check := adsFixture(t)
			store := &ADSStore{}
			_, getErr := store.Get(path, name)
			check()
			setErr := store.Set(path, name, []byte("attribute"))
			check()
			removeErr := store.Remove(path, name)
			check()
			for op, err := range map[string]error{"Get": getErr, "Set": setErr, "Remove": removeErr} {
				var errno syscall.Errno
				if !errors.As(err, &errno) || errno == 0 {
					t.Errorf("%s: lost native error: %v", op, err)
				}
			}
		})
	}
}

func TestADSListErrors(t *testing.T) {
	path, check := adsFixture(t)
	store := &ADSStore{}
	for _, missing := range []string{path + ".missing", filepath.Join(path, "child")} {
		names, err := store.List(missing)
		check()
		if !errors.Is(err, os.ErrNotExist) || len(names) != 0 {
			t.Errorf("List(%q): got %q, %v; want not-exist", missing, names, err)
		}
	}
	if _, err := store.List(path + "\x00"); err == nil {
		t.Error("List accepted NUL in path")
	}
	check()
}

func TestADSListEmptyFileAndDirectory(t *testing.T) {
	path, check := adsFixture(t)
	store := &ADSStore{}
	for _, target := range []string{path, t.TempDir()} {
		names, err := store.List(target)
		check()
		if err != nil || len(names) != 0 {
			t.Errorf("List(%q): got %q, %v; want empty success", target, names, err)
		}
	}
}

func TestADSListNextErrorDiscardsPartialResult(t *testing.T) {
	path, check := adsFixture(t)
	store := &ADSStore{}
	err := store.Set(path, "user.comment", []byte("retained attribute"))
	check()
	if err != nil {
		t.Fatal(err)
	}
	// Native enumeration normally ends at EOF. Inject a continuation failure
	// after a real named-stream result, retaining the real handle for FindClose.
	nativeNext := findNextStreamW
	t.Cleanup(func() { findNextStreamW = nativeNext })
	calls := 0
	findNextStreamW = func(handle uintptr, data *WIN32_FIND_STREAM_DATA) (uintptr, error) {
		calls++
		if calls == 2 {
			return 0, syscall.ERROR_ACCESS_DENIED
		}
		return nativeNext(handle, data)
	}
	names, err := store.List(path)
	check()
	if !errors.Is(err, syscall.ERROR_ACCESS_DENIED) || len(names) != 0 {
		t.Errorf("List hid continuation failure or returned a partial result: %q, %v", names, err)
	}
	got, err := store.Get(path, "user.comment")
	check()
	if err != nil || string(got) != "retained attribute" {
		t.Errorf("enumeration failure changed attribute: %q, %v", got, err)
	}
}

func TestADSDirectoryRoundTrip(t *testing.T) {
	path, check := adsFixture(t)
	dir := filepath.Dir(path)
	store := &ADSStore{}
	if err := store.Set(dir, "user.comment", []byte("directory attribute")); err != nil {
		t.Fatal(err)
	}
	check()
	got, err := store.Get(dir, "user.comment")
	check()
	if err != nil || string(got) != "directory attribute" {
		t.Fatalf("Get directory attribute: %q, %v", got, err)
	}
	names, err := store.List(dir)
	check()
	if err != nil || !slices.Equal(names, []string{"user.comment"}) {
		t.Fatalf("List directory: %q, %v", names, err)
	}
	err = store.Remove(dir, "user.comment")
	check()
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("directory did not survive: %v", err)
	}
	names, err = store.List(dir)
	check()
	if err != nil || len(names) != 0 {
		t.Errorf("List directory after removal: %q, %v", names, err)
	}
}

func TestADSDirectoryDefaultStream(t *testing.T) {
	for _, name := range []string{":$INDEX_ALLOCATION", "$I30:$INDEX_ALLOCATION", "$i30:$index_allocation", "$I30:$INDEX_ALLOCATION ", "$I30:$INDEX_ALLOCATION."} {
		for _, op := range []string{"Get", "Set", "Remove"} {
			t.Run(fmt.Sprintf("%s/%q", op, name), func(t *testing.T) {
				path, check := adsFixture(t)
				dir := filepath.Join(filepath.Dir(path), "empty")
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
				checkDir := func() {
					t.Helper()
					if info, err := os.Stat(dir); err != nil || !info.IsDir() {
						t.Errorf("original directory did not survive: %v", err)
					}
				}
				t.Cleanup(checkDir)
				store := &ADSStore{}
				var err error
				switch op {
				case "Get":
					_, err = store.Get(dir, name)
				case "Set":
					err = store.Set(dir, name, []byte("attribute"))
				case "Remove":
					err = store.Remove(dir, name)
				}
				check()
				checkDir()
				if err == nil {
					t.Error("directory default stream accepted as an attribute")
				}
			})
		}
	}
}

func TestADSOrdinaryDotSpaceNames(t *testing.T) {
	for _, name := range []string{".", "..", " ", ". ", "..."} {
		for _, op := range []string{"Get", "Set", "Remove"} {
			t.Run(fmt.Sprintf("%s/%q", op, name), func(t *testing.T) {
				path, check := adsFixture(t)
				store := &ADSStore{}
				var err error
				switch op {
				case "Get":
					var value []byte
					value, err = store.Get(path, name)
					if len(value) != 0 {
						t.Errorf("Get exposed default-stream contents: %q", value)
					}
				case "Set":
					err = store.Set(path, name, []byte("attribute"))
				case "Remove":
					err = store.Remove(path, name)
				}
				check()
				var errno syscall.Errno
				if !errors.As(err, &errno) || errno == 0 {
					t.Errorf("ordinary path: want native name error, got %v", err)
				}
			})
		}
	}
}

func TestADSDotSpaceNamesInBasePath(t *testing.T) {
	for _, relative := range []bool{false, true} {
		for _, name := range []string{".", "..", " ", ". ", "..."} {
			t.Run(fmt.Sprintf("relative=%t/%q", relative, name), func(t *testing.T) {
				path, check := adsFixture(t)
				target := path + ":" + name
				if relative {
					t.Chdir(filepath.Dir(path))
					target = filepath.Base(path) + ":" + name
				}
				store := &ADSStore{}
				value := []byte("named attribute")
				err := store.Set(target, "$DATA", value)
				check()
				if err != nil {
					t.Fatalf("Set explicitly typed native stream: %v", err)
				}
				got, err := store.Get(target, "$DATA")
				check()
				if err != nil || !bytes.Equal(got, value) {
					t.Fatalf("Get explicitly typed native stream: %q, %v", got, err)
				}
				names, err := store.List(path)
				check()
				if err != nil || !slices.Equal(names, []string{name}) {
					t.Errorf("List explicitly typed native stream: %q, %v", names, err)
				}
				err = store.Remove(target, "$DATA")
				check()
				if err != nil {
					t.Fatal(err)
				}
				names, err = store.List(path)
				check()
				if err != nil || len(names) != 0 {
					t.Errorf("List after removal: %q, %v", names, err)
				}
			})
		}
	}
}

func TestADSRelativeSingleCharacterFile(t *testing.T) {
	path, check := adsNamedFixture(t, "a")
	t.Chdir(filepath.Dir(path))
	store := &ADSStore{}
	value := []byte("relative file attribute")
	err := store.Set("a", "user.comment", value)
	check()
	if err != nil {
		t.Fatalf("Set relative one-character file: %v", err)
	}
	got, err := os.ReadFile(path + ":user.comment")
	check()
	if err != nil || !bytes.Equal(got, value) {
		t.Fatalf("attribute did not reach the intended native stream: %q, %v", got, err)
	}
	got, err = store.Get("a", "user.comment")
	check()
	if err != nil || !bytes.Equal(got, value) {
		t.Fatalf("Get relative file: %q, %v", got, err)
	}
	names, err := store.List("a")
	check()
	if err != nil || !slices.Equal(names, []string{"user.comment"}) {
		t.Fatalf("List relative file: %q, %v", names, err)
	}
	err = store.Remove("a", "user.comment")
	check()
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.ReadFile(path + ":user.comment")
	check()
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("relative Remove left native stream: %v", err)
	}
}
