// Copyright (c) 2026 GNOMATIX. All rights reserved.

package fallback

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

// RegistryStore implements a simple JSON-based fallback for attributes.
// This is used when native xattr/ADS are not available.
type RegistryStore struct {
	path string
}

type registry map[string]map[string][]byte

func NewRegistryStore() (*RegistryStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".gnomatix", "xyzattr")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	s := &RegistryStore{
		path: filepath.Join(dir, "attributes.json"),
	}
	if _, err := s.read(); err != nil {
		return nil, err
	}
	return s, nil
}

// load is called while holding the registry lock. Only a missing file is an
// empty registry; unreadable or malformed storage must never be overwritten.
func (s *RegistryStore) load() (registry, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return make(registry), nil
	}
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("decode registry %q: invalid UTF-8: %w", s.path, os.ErrInvalid)
	}
	if err := validateUnicodeEscapes(data); err != nil {
		return nil, fmt.Errorf("decode registry %q: %w", s.path, err)
	}
	var db registry
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("decode registry %q: %w", s.path, err)
	}
	if db == nil {
		return nil, fmt.Errorf("decode registry %q: expected JSON object, got null", s.path)
	}
	return db, nil
}

// validateUnicodeEscapes rejects the unpaired surrogate escapes that Unmarshal
// otherwise silently replaces. JSON syntax stays with encoding/json: malformed
// documents pass through to Unmarshal for its normal SyntaxError. In valid JSON,
// backslashes only occur in strings, so walking whole escapes needs no structural
// parser and distinguishes a literal backslash from a Unicode escape.
func validateUnicodeEscapes(data []byte) error {
	if !bytes.Contains(data, []byte(`\u`)) || !json.Valid(data) {
		return nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] != '\\' {
			continue
		}
		first, ok := unicodeEscape(data[i:])
		if !ok {
			i++ // Consume an ordinary escape, including an escaped backslash.
			continue
		}
		if !utf16.IsSurrogate(first) {
			i += 5
			continue
		}
		second, ok := unicodeEscape(data[i+6:])
		if !ok || utf16.DecodeRune(first, second) == utf8.RuneError {
			return fmt.Errorf("unpaired UTF-16 surrogate at byte %d: %w", i, os.ErrInvalid)
		}
		i += 11 // Consume both six-byte escapes of a valid surrogate pair.
	}
	return nil
}

func unicodeEscape(data []byte) (rune, bool) {
	if len(data) < 6 || data[0] != '\\' || data[1] != 'u' {
		return 0, false
	}
	value, err := strconv.ParseUint(string(data[2:6]), 16, 16)
	return rune(value), err == nil
}

// The lock file is separate from attributes.json because replacing the data
// file changes its identity. Never unlink the lock file: other processes may
// already have it open. Closing the handle releases its OS lock, including on
// process exit.
func (s *RegistryStore) lock(exclusive bool) (*os.File, error) {
	f, err := os.OpenFile(s.path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := lockFile(f, exclusive); err != nil {
		f.Close()
		return nil, fmt.Errorf("lock registry %q: %w", s.path, err)
	}
	return f, nil
}

func (s *RegistryStore) read() (registry, error) {
	lock, err := s.lock(false)
	if err != nil {
		return nil, err
	}
	defer lock.Close()
	return s.load()
}

func (s *RegistryStore) update(change func(registry) bool) error {
	lock, err := s.lock(true)
	if err != nil {
		return err
	}
	defer lock.Close()
	db, err := s.load()
	if err != nil {
		return err
	}
	if !change(db) {
		return nil
	}
	return s.save(db)
}

// save publishes only a completely written and synced file, in the same
// directory. The lock covers read/modify/replace across instances and processes.
// No mutable cache survives a failed commit. Power-loss durability of directory
// entries remains subject to the filesystem's rename semantics.
func (s *RegistryStore) save(db registry) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".attributes-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.path)
}

func (s *RegistryStore) Get(path string, name string) ([]byte, error) {
	if err := validateKeys(path, name); err != nil {
		return nil, err
	}
	db, err := s.read()
	if err != nil {
		return nil, err
	}
	if attrs, ok := db[path]; ok {
		if val, ok := attrs[name]; ok {
			return val, nil
		}
	}
	return nil, os.ErrNotExist
}

func (s *RegistryStore) Set(path string, name string, value []byte) error {
	if err := validateKeys(path, name); err != nil {
		return err
	}
	return s.update(func(db registry) bool {
		if db[path] == nil {
			db[path] = make(map[string][]byte)
		}
		db[path][name] = value
		return true
	})
}

func (s *RegistryStore) List(path string) ([]string, error) {
	if err := validateKeys(path, ""); err != nil {
		return nil, err
	}
	db, err := s.read()
	if err != nil {
		return nil, err
	}
	var names []string
	if attrs, ok := db[path]; ok {
		for name := range attrs {
			names = append(names, name)
		}
	}
	return names, nil
}

func (s *RegistryStore) Remove(path string, name string) error {
	if err := validateKeys(path, name); err != nil {
		return err
	}
	return s.update(func(db registry) bool {
		if attrs, ok := db[path]; ok {
			if _, exists := attrs[name]; !exists {
				return false
			}
			delete(attrs, name)
			if len(attrs) == 0 {
				delete(db, path)
			}
			return true
		}
		return false
	})
}

// JSON object keys cannot preserve invalid UTF-8 bytes. Fail explicitly rather
// than silently persisting a different path or attribute name. This restriction
// belongs to this registry format; it does not constrain native engines or
// binary attribute values, and does not introduce an alternative key encoding.
func validateKeys(path, name string) error {
	if !utf8.ValidString(path) {
		return fmt.Errorf("registry path is not valid UTF-8: %w", os.ErrInvalid)
	}
	if !utf8.ValidString(name) {
		return fmt.Errorf("registry attribute name is not valid UTF-8: %w", os.ErrInvalid)
	}
	return nil
}
