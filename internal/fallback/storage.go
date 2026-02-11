// Copyright (c) 2026 GNOMATIX. All rights reserved.

package fallback

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// RegistryStore implements a simple JSON-based fallback for attributes.
// This is used when native xattr/ADS are not available.
type RegistryStore struct {
	mu   sync.RWMutex
	path string
	db   map[string]map[string][]byte
}

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
		db:   make(map[string]map[string][]byte),
	}
	s.load()
	return s, nil
}

func (s *RegistryStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &s.db)
}

func (s *RegistryStore) save() error {
	data, err := json.MarshalIndent(s.db, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600)
}

func (s *RegistryStore) Get(path string, name string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if attrs, ok := s.db[path]; ok {
		if val, ok := attrs[name]; ok {
			return val, nil
		}
	}
	return nil, os.ErrNotExist
}

func (s *RegistryStore) Set(path string, name string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db[path] == nil {
		s.db[path] = make(map[string][]byte)
	}
	s.db[path][name] = value
	return s.save()
}

func (s *RegistryStore) List(path string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var names []string
	if attrs, ok := s.db[path]; ok {
		for name := range attrs {
			names = append(names, name)
		}
	}
	return names, nil
}

func (s *RegistryStore) Remove(path string, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if attrs, ok := s.db[path]; ok {
		delete(attrs, name)
		if len(attrs) == 0 {
			delete(s.db, path)
		}
		return s.save()
	}
	return nil
}
