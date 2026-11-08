// Copyright (c) 2026 GNOMATIX. All rights reserved.

package xyzattr

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gitea.gnomatix.com/gnomatix/xyzattr/internal/fs"
)

func TestFallbackInitializationChild(t *testing.T) {
	if os.Getenv("XYZATTR_INIT_CHILD") != "1" {
		return
	}
	// An invalid path exercises the existing Unknown-filesystem selection. The
	// naming/detection fallback policy itself is outside this regression.
	store := At("\x00")
	checks := []struct {
		name string
		call func() error
	}{
		{"get", func() error { _, err := store.Get("name"); return err }},
		{"set", func() error { return store.Set("name", []byte("value")) }},
		{"list", func() error { _, err := store.List(); return err }},
		{"remove", func() error { return store.Remove("name") }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("initialization failure panicked: %v", p)
				}
			}()
			err := check.call()
			if err == nil {
				t.Fatal("initialization failure hidden")
			}
			switch os.Getenv("XYZATTR_INIT_KIND") {
			case "surrogate-path", "surrogate-name":
				if !errors.Is(err, os.ErrInvalid) {
					t.Fatalf("Unicode validation cause lost: %v", err)
				}
			case "mkdir", "read":
				var pathErr *os.PathError
				if !errors.As(err, &pathErr) {
					t.Fatalf("initialization cause lost: %v", err)
				}
			case "decode":
				var syntaxErr *json.SyntaxError
				if !errors.As(err, &syntaxErr) {
					t.Fatalf("decode cause lost: %v", err)
				}
			}
		})
	}
	path := filepath.Join(t.TempDir(), "native.txt")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	fsType, err := fs.Detect(path)
	if err != nil {
		t.Fatal(err)
	}
	if getNativeEngineByFS(fsType) != nil {
		native := At(path)
		if err := native.Set("user.init-independent", []byte("value")); err != nil {
			t.Fatalf("registry init blocked native Set: %v", err)
		}
		value, err := native.Get("user.init-independent")
		if err != nil || string(value) != "value" {
			t.Fatalf("native Get = %q, %v", value, err)
		}
		if err := native.Remove("user.init-independent"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFallbackInitializationErrorsReachAPI(t *testing.T) {
	for _, kind := range []string{"mkdir", "decode", "read", "surrogate-path", "surrogate-name"} {
		t.Run(kind, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(home, ".gnomatix", "xyzattr", "attributes.json")
			source := "{"
			if kind == "surrogate-path" {
				source = `{"\ud800":{"a":"c2F2ZWQ="}}`
			}
			if kind == "surrogate-name" {
				source = `{"p":{"\udfff":"c2F2ZWQ="}}`
			}
			if kind == "mkdir" {
				if err := os.WriteFile(filepath.Join(home, ".gnomatix"), []byte("sentinel"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if kind == "read" {
					if err := os.Mkdir(path, 0700); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(path, []byte(source), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestFallbackInitializationChild$", "-test.v")
			for _, e := range os.Environ() {
				key := strings.SplitN(e, "=", 2)[0]
				if !strings.EqualFold(key, "HOME") && !strings.EqualFold(key, "USERPROFILE") {
					cmd.Env = append(cmd.Env, e)
				}
			}
			cmd.Env = append(cmd.Env, "HOME="+home, "USERPROFILE="+home, "XYZATTR_INIT_CHILD=1", "XYZATTR_INIT_KIND="+kind)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("initialization child: %v\n%s", err, output)
			}
			if kind != "mkdir" && kind != "read" {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != source {
					t.Fatalf("invalid registry changed: %q, %v", data, err)
				}
			}
		})
	}
}
