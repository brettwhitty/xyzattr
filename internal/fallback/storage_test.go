// Copyright (c) 2026 GNOMATIX. All rights reserved.

package fallback

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

func fixtureHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return filepath.Join(home, ".gnomatix", "xyzattr", "attributes.json")
}

func openStore(t *testing.T) *RegistryStore {
	t.Helper()
	s, err := NewRegistryStore()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func requireValue(t *testing.T, s *RegistryStore, path, name string, want []byte) {
	t.Helper()
	got, err := s.Get(path, name)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("Get(%q, %q) = %x, %v; want %x", path, name, got, err, want)
	}
}

func TestRegistryReopenAndOwnership(t *testing.T) {
	registry := fixtureHome(t)
	s := openStore(t)
	path := "./literal/../path"
	value := []byte{0, 255, 128, 1}
	if err := s.Set(path, "name", value); err != nil {
		t.Fatal(err)
	}
	value[0] = 42
	requireValue(t, s, path, "name", []byte{0, 255, 128, 1})
	got, err := s.Get(path, "name")
	if err != nil {
		t.Fatal(err)
	}
	got[1] = 42
	requireValue(t, s, path, "name", []byte{0, 255, 128, 1})
	for name, value := range map[string][]byte{"empty": {}, "nil": nil} {
		if err := s.Set(path, name, value); err != nil {
			t.Fatal(err)
		}
	}
	reopened := openStore(t)
	requireValue(t, reopened, path, "name", []byte{0, 255, 128, 1})
	if _, err := reopened.Get("path", "name"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("path keys changed: %v", err)
	}
	names, err := reopened.List(path)
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(names)
	if !reflect.DeepEqual(names, []string{"empty", "name", "nil"}) {
		t.Fatalf("names = %v", names)
	}
	data, err := os.ReadFile(registry)
	if err != nil {
		t.Fatal(err)
	}
	var disk map[string]map[string][]byte
	if err := json.Unmarshal(data, &disk); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(disk[path]["name"], []byte{0, 255, 128, 1}) {
		t.Fatalf("disk = %s", data)
	}
	if err := reopened.Remove(path, "name"); err != nil {
		t.Fatal(err)
	}
	if _, err := openStore(t).Get(path, "name"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("removed value: %v", err)
	}
}

func TestRegistryExistingFormatAndPathKeys(t *testing.T) {
	path := fixtureHome(t)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	const legacy = `{"./a/../b":{"binary":"AP8BgA==","nil":null,"empty":""},"unrelated":{"saved":"eWVz"}}`
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatal(err)
	}
	s := openStore(t)
	requireValue(t, s, "./a/../b", "binary", []byte{0, 255, 1, 128})
	requireValue(t, s, "./a/../b", "nil", nil)
	requireValue(t, s, "./a/../b", "empty", []byte{})
	if err := s.Set("./a/../b", "added", []byte("new")); err != nil {
		t.Fatal(err)
	}
	requireValue(t, openStore(t), "unrelated", "saved", []byte("yes"))
	if _, err := s.Get("b", "binary"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("path key normalized: %v", err)
	}
}

func TestRegistryRejectsLossyKeys(t *testing.T) {
	for _, key := range []string{"path", "name"} {
		for _, operation := range []string{"get", "set", "list", "remove"} {
			if key == "name" && operation == "list" {
				continue
			}
			t.Run(key+"/"+operation, func(t *testing.T) {
				file := fixtureHome(t)
				s := openStore(t)
				if err := s.Set("preserved", "name", []byte("saved")); err != nil {
					t.Fatal(err)
				}
				before, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				path, name := "path", "name"
				if key == "path" {
					path = string([]byte{'p', 0xff})
				} else {
					name = string([]byte{'a', 0xff})
				}
				switch operation {
				case "get":
					_, err = s.Get(path, name)
				case "set":
					err = s.Set(path, name, []byte{0, 255})
				case "list":
					_, err = s.List(path)
				case "remove":
					err = s.Remove(path, name)
				}
				if !errors.Is(err, os.ErrInvalid) {
					t.Errorf("error = %v; want explicit invalid-key error", err)
				}
				after, err := os.ReadFile(file)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatalf("registry changed on invalid key: %q, %v", after, err)
				}
			})
		}
	}
}

func TestRegistryRejectsInvalidUTF8Storage(t *testing.T) {
	file := fixtureHome(t)
	s := openStore(t)
	data := append([]byte(`{"p`), 0xff)
	data = append(data, []byte(`":{"a":"c2F2ZWQ="}}`)...)
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewRegistryStore(); err == nil {
		t.Error("constructor accepted invalid UTF-8 JSON")
	}
	if err := s.Set("valid", "a", []byte("new")); err == nil {
		t.Error("Set rewrote invalid UTF-8 JSON")
	}
	after, err := os.ReadFile(file)
	if err != nil || !bytes.Equal(data, after) {
		t.Fatalf("invalid source changed: %q, %v", after, err)
	}
}

func TestRegistryRejectsUnpairedSurrogates(t *testing.T) {
	for _, tc := range []struct{ name, source string }{
		{"high-path", `{"\ud800":{"a":"c2F2ZWQ="},"unrelated":{"a":"c2F2ZWQ="}}`},
		{"low-path", `{"\udfff":{"a":"c2F2ZWQ="},"unrelated":{"a":"c2F2ZWQ="}}`},
		{"high-name", `{"p":{"\ud800":"c2F2ZWQ="},"unrelated":{"a":"c2F2ZWQ="}}`},
		{"low-name", `{"p":{"\udfff":"c2F2ZWQ="},"unrelated":{"a":"c2F2ZWQ="}}`},
		{"high-nonlow", `{"\ud800\u0061":{"a":"c2F2ZWQ="},"unrelated":{"a":"c2F2ZWQ="}}`},
		{"reversed-pair", `{"p":{"\udfff\ud800":"c2F2ZWQ="},"unrelated":{"a":"c2F2ZWQ="}}`},
	} {
		for _, operation := range []string{"initialize", "get", "list", "set", "remove"} {
			t.Run(tc.name+"/"+operation, func(t *testing.T) {
				file := fixtureHome(t)
				s := openStore(t)
				if err := os.WriteFile(file, []byte(tc.source), 0600); err != nil {
					t.Fatal(err)
				}
				var err error
				switch operation {
				case "initialize":
					var reopened *RegistryStore
					reopened, err = NewRegistryStore()
					if reopened != nil {
						t.Error("constructor returned a store for lossy JSON")
					}
				case "get":
					_, err = s.Get("unrelated", "a")
				case "list":
					_, err = s.List("unrelated")
				case "set":
					err = s.Set("unrelated", "new", []byte("new"))
				case "remove":
					err = s.Remove("unrelated", "a")
				}
				if !errors.Is(err, os.ErrInvalid) {
					t.Errorf("%s error = %v; want unrepresentable Unicode error", operation, err)
				}
				after, readErr := os.ReadFile(file)
				if readErr != nil || string(after) != tc.source {
					t.Fatalf("offending source rewritten: %q, %v", after, readErr)
				}
			})
		}
	}
}

func TestRegistrySupportedUnicodeEscapes(t *testing.T) {
	for _, tc := range []struct{ name, rawPath, rawName, path, attr string }{
		{"pair", `"\ud83d\ude00"`, `"\uD83D\uDE01"`, "😀", "😁"},
		{"ordinary-escapes", `"x\"\\\/\b\f\n\r\t"`, `"a\"\\\/\b\f\n\r\t"`, "x\"\\/\b\f\n\r\t", "a\"\\/\b\f\n\r\t"},
		{"literal-backslashes", `"\\ud800"`, `"\\udfff"`, `\ud800`, `\udfff`},
		{"unicode-backslashes", `"\u005cud800"`, `"\u005cudfff"`, `\ud800`, `\udfff`},
		{"backslash-and-pair", `"\\\ud83d\ude00"`, `"a"`, "\\😀", "a"},
		{"replacement-character", `"�"`, `"\ufffd"`, "�", "�"},
		{"non-ASCII", `"é/日本"`, `"π"`, "é/日本", "π"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := fixtureHome(t)
			if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
				t.Fatal(err)
			}
			source := fmt.Sprintf(`{%s:{%s:"c2F2ZWQ="}}`, tc.rawPath, tc.rawName)
			if err := os.WriteFile(file, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			s := openStore(t)
			requireValue(t, s, tc.path, tc.attr, []byte("saved"))
			names, err := s.List(tc.path)
			if err != nil || !reflect.DeepEqual(names, []string{tc.attr}) {
				t.Fatalf("List = %q, %v", names, err)
			}
			if err := s.Set("unrelated", "a", []byte("new")); err != nil {
				t.Fatal(err)
			}
			requireValue(t, openStore(t), tc.path, tc.attr, []byte("saved"))
			if err := s.Remove(tc.path, tc.attr); err != nil {
				t.Fatal(err)
			}
			if _, err := openStore(t).Get(tc.path, tc.attr); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("removed attribute = %v", err)
			}
		})
	}
}

func TestRegistryUnicodeSyntaxErrorsRemainJSONErrors(t *testing.T) {
	file := fixtureHome(t)
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		t.Fatal(err)
	}
	const source = `{"\ud8":{"a":"c2F2ZWQ="}}`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := NewRegistryStore()
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("error = %v; want Go JSON syntax error", err)
	}
	after, readErr := os.ReadFile(file)
	if readErr != nil || string(after) != source {
		t.Fatalf("source changed: %q, %v", after, readErr)
	}
}

func TestRegistryRejectsInvalidStorage(t *testing.T) {
	for _, data := range []string{"", "{", "null", "[]", `{"p":{"a":"not base64!"}}`, `{"p":{"a":"b2s="}} trailing`} {
		t.Run(fmt.Sprintf("%q", data), func(t *testing.T) {
			path := fixtureHome(t)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			s, err := NewRegistryStore()
			if err == nil || s != nil {
				t.Errorf("constructor = %v, %v; want nil, error", s, err)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != data {
				t.Fatalf("source changed: %q, %v", got, err)
			}
		})
	}
	t.Run("read error", func(t *testing.T) {
		path := fixtureHome(t)
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
		if s, err := NewRegistryStore(); err == nil || s != nil {
			t.Fatalf("constructor = %v, %v; want read error", s, err)
		}
	})
}

func TestRegistryDoesNotHideLaterLoadFailures(t *testing.T) {
	for _, operation := range []string{"get", "list", "set", "remove"} {
		t.Run(operation, func(t *testing.T) {
			path := fixtureHome(t)
			s := openStore(t)
			if err := s.Set("p", "a", []byte("saved")); err != nil {
				t.Fatal(err)
			}
			bad := []byte(`{"damaged":`)
			if err := os.WriteFile(path, bad, 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			switch operation {
			case "get":
				_, err = s.Get("p", "a")
			case "list":
				_, err = s.List("p")
			case "set":
				err = s.Set("p", "b", []byte("new"))
			case "remove":
				err = s.Remove("p", "a")
			}
			if err == nil {
				t.Error("load failure was hidden")
			}
			got, readErr := os.ReadFile(path)
			if readErr != nil || !bytes.Equal(got, bad) {
				t.Fatalf("damaged source overwritten: %s, %v", got, readErr)
			}
		})
	}
}

func TestRegistryFailedSaveIsNotVisible(t *testing.T) {
	for _, operation := range []string{"set", "remove"} {
		t.Run(operation, func(t *testing.T) {
			path := fixtureHome(t)
			s := openStore(t)
			if err := s.Set("p", "a", []byte("saved")); err != nil {
				t.Fatal(err)
			}
			backup := path + ".saved"
			if err := os.Rename(path, backup); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}
			var err error
			if operation == "set" {
				err = s.Set("p", "a", []byte("failed"))
			} else {
				err = s.Remove("p", "a")
			}
			if err == nil {
				t.Error("save failure hidden")
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(backup, path); err != nil {
				t.Fatal(err)
			}
			requireValue(t, s, "p", "a", []byte("saved"))
			requireValue(t, openStore(t), "p", "a", []byte("saved"))
		})
	}
}

func TestRegistryIndependentInstancesMerge(t *testing.T) {
	fixtureHome(t)
	a, b := openStore(t), openStore(t)
	if err := a.Set("p", "first", []byte("1")); err != nil {
		t.Fatal(err)
	}
	if err := b.Set("q", "second", []byte("2")); err != nil {
		t.Fatal(err)
	}
	requireValue(t, a, "q", "second", []byte("2"))
	requireValue(t, b, "p", "first", []byte("1"))
	if err := a.Remove("p", "first"); err != nil {
		t.Fatal(err)
	}
	if err := b.Set("q", "third", []byte("3")); err != nil {
		t.Fatal(err)
	}
	if _, err := openStore(t).Get("p", "first"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("removed key resurrected: %v", err)
	}
	requireValue(t, openStore(t), "q", "second", []byte("2"))
}

func TestRegistryProcessWriter(t *testing.T) {
	id := os.Getenv("XYZATTR_TEST_WRITER")
	if id == "" {
		return
	}
	s := openStore(t)
	fmt.Println("ready")
	if _, err := bufio.NewReader(os.Stdin).ReadByte(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		if err := s.Set("shared-path", fmt.Sprintf("%s-%d", id, i), []byte("preserved")); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRegistryConcurrentProcesses(t *testing.T) {
	fixtureHome(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	type child struct {
		cmd    *exec.Cmd
		stdin  io.WriteCloser
		output *bufio.Reader
		stderr bytes.Buffer
	}
	children := make([]*child, 6)
	for i := range children {
		c := &child{cmd: exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRegistryProcessWriter$")}
		children[i] = c
		c.cmd.Env = append(os.Environ(), fmt.Sprintf("XYZATTR_TEST_WRITER=%d", i))
		c.cmd.Stderr = &c.stderr
		var err error
		c.stdin, err = c.cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		out, err := c.cmd.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		c.output = bufio.NewReader(out)
		if err := c.cmd.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if c.cmd.ProcessState == nil {
				c.cmd.Process.Kill()
				c.cmd.Wait()
			}
		})
		line, err := c.output.ReadString('\n')
		if err != nil || line != "ready\n" {
			t.Fatalf("child ready: %q, %v, %s", line, err, &c.stderr)
		}
	}
	for _, c := range children {
		if _, err := c.stdin.Write([]byte{1}); err != nil {
			t.Fatal(err)
		}
		c.stdin.Close()
	}
	for _, c := range children {
		output, _ := io.ReadAll(c.output)
		if err := c.cmd.Wait(); err != nil {
			t.Fatalf("child: %v\n%s\n%s", err, output, &c.stderr)
		}
	}
	s := openStore(t)
	for i := range children {
		for j := 0; j < 12; j++ {
			requireValue(t, s, "shared-path", fmt.Sprintf("%d-%d", i, j), []byte("preserved"))
		}
	}
}

func TestRegistryTargetIsUnchanged(t *testing.T) {
	fixtureHome(t)
	s := openStore(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "target")
	original := []byte{0, 255, 1, 128}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Set(path, "user.name", []byte("value")); err != nil {
		t.Fatal(err)
	}
	requireValue(t, openStore(t), path, "user.name", []byte("value"))
	if err := s.Remove(path, "user.name"); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, original) || !os.SameFile(before, after) {
		t.Fatalf("target changed: %x, %v", data, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "target" {
		t.Fatalf("target directory changed: %v, %v", entries, err)
	}
}

func TestRegistryLockFailurePreservesData(t *testing.T) {
	path := fixtureHome(t)
	s := openStore(t)
	if err := s.Set("p", "a", []byte("saved")); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path + ".lock"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path+".lock", 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := NewRegistryStore(); err == nil {
		t.Error("constructor ignored lock failure")
	}
	if _, err := s.Get("p", "a"); err == nil {
		t.Error("Get ignored lock failure")
	}
	if _, err := s.List("p"); err == nil {
		t.Error("List ignored lock failure")
	}
	if err := s.Set("p", "a", []byte("failed")); err == nil {
		t.Error("Set ignored lock failure")
	}
	if err := s.Remove("p", "a"); err == nil {
		t.Error("Remove ignored lock failure")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("committed bytes changed: %q, %v", after, err)
	}
}
