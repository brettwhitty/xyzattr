// Copyright (c) 2026 GNOMATIX. All rights reserved.

package fallback

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRegistryAtomicReplacement(t *testing.T) {
	path := fixtureHome(t)
	s := openStore(t)
	if err := s.Set("p", "a", []byte("old")); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := s.Set("p", "a", []byte("new")); err != nil {
		t.Fatal(err)
	}
	held, err := io.ReadAll(f)
	if err != nil || !bytes.Equal(before, held) {
		t.Fatalf("existing registry was overwritten in place: %q, %v", held, err)
	}
	requireValue(t, s, "p", "a", []byte("new"))
}

func TestRegistryWriteFailureChild(t *testing.T) {
	if os.Getenv("XYZATTR_LIMIT_CHILD") != "1" {
		return
	}
	s := openStore(t)
	before, err := os.ReadFile(s.path)
	if err != nil {
		t.Fatal(err)
	}
	signal.Ignore(syscall.SIGXFSZ)
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &limit); err != nil {
		t.Fatal(err)
	}
	limit.Cur = 512
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit); err != nil {
		t.Fatal(err)
	}
	err = s.Set("p", "a", bytes.Repeat([]byte("x"), 4096))
	if !errors.Is(err, syscall.EFBIG) {
		t.Fatalf("Set error = %v; want EFBIG", err)
	}
	after, readErr := os.ReadFile(s.path)
	if readErr != nil || !bytes.Equal(before, after) {
		t.Errorf("failed write damaged registry: %q, %v", after, readErr)
	}
	requireValue(t, s, "p", "a", []byte("saved"))
	requireValue(t, openStore(t), "p", "a", []byte("saved"))
}

func TestRegistryPartialWritePreservesCommittedData(t *testing.T) {
	fixtureHome(t)
	s := openStore(t)
	if err := s.Set("p", "a", []byte("saved")); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRegistryWriteFailureChild$", "-test.v")
	cmd.Env = append(os.Environ(), "XYZATTR_LIMIT_CHILD=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("write failure child: %v\n%s", err, output)
	}
}

func TestRegistryPermissionFailuresPreserveData(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission denial requires an unprivileged fixture process")
	}
	for _, kind := range []string{"read", "create"} {
		t.Run(kind, func(t *testing.T) {
			path := fixtureHome(t)
			s := openStore(t)
			if err := s.Set("p", "a", []byte("saved")); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			blocked, mode := path, os.FileMode(0000)
			if kind == "create" {
				blocked, mode = filepath.Dir(path), 0500
			}
			if err := os.Chmod(blocked, mode); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.Chmod(blocked, 0700) })
			if kind == "read" {
				if _, err := NewRegistryStore(); !errors.Is(err, os.ErrPermission) {
					t.Fatalf("constructor = %v; want permission error", err)
				}
				if _, err := s.Get("p", "a"); !errors.Is(err, os.ErrPermission) {
					t.Fatalf("Get = %v", err)
				}
				if _, err := s.List("p"); !errors.Is(err, os.ErrPermission) {
					t.Fatalf("List = %v", err)
				}
			}
			if err := s.Set("p", "a", []byte("failed")); !errors.Is(err, os.ErrPermission) {
				t.Fatalf("Set = %v", err)
			}
			if err := s.Remove("p", "a"); !errors.Is(err, os.ErrPermission) {
				t.Fatalf("Remove = %v", err)
			}
			if err := os.Chmod(blocked, 0700); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("committed bytes changed: %q, %v", after, err)
			}
			requireValue(t, s, "p", "a", []byte("saved"))
			requireValue(t, openStore(t), "p", "a", []byte("saved"))
		})
	}
}

func TestRegistryRejectsUnrepresentableNativePath(t *testing.T) {
	file := fixtureHome(t)
	s := openStore(t)
	path := filepath.Join(t.TempDir(), string([]byte{'p', 0xff}))
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(path, "a", []byte("saved")); !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("Set = %v; want explicit representation error", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "original" {
		t.Fatalf("native file changed: %q, %v", data, err)
	}
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("invalid key created registry data: %v", err)
	}
}
