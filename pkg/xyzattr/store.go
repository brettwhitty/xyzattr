// Copyright (c) 2026 GNOMATIX. All rights reserved.

package xyzattr

import (
	"errors"
	"os"
	"sync"

	"gitea.gnomatix.com/gnomatix/xyzattr/internal/fallback"
	"gitea.gnomatix.com/gnomatix/xyzattr/internal/fs"
)

// Engine defines the internal interface for platform/filesystem-specific storage.
type Engine interface {
	Get(path string, name string) ([]byte, error)
	Set(path string, name string, value []byte) error
	List(path string) ([]string, error)
	Remove(path string, name string) error
}

// FileStore provides a fluent API for attribute operations on a specific file.
type FileStore struct {
	path   string
	engine Engine
}

var (
	cacheMu sync.RWMutex
	fsCache = make(map[fs.Type]Engine)
	reg     Engine
)

func init() {
	var err error
	reg, err = fallback.NewRegistryStore()
	if err != nil {
		// Proceed with registry as nil; native will still be tried.
	}
}

// At returns a FileStore for the given path, caching the best engine for the filesystem.
func At(path string) *FileStore {
	return &FileStore{
		path:   path,
		engine: resolveEngine(path),
	}
}

func resolveEngine(path string) Engine {
	t, _ := fs.Detect(path)

	cacheMu.RLock()
	cached, ok := fsCache[t]
	cacheMu.RUnlock()
	if ok {
		return cached
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double check
	if cached, ok := fsCache[t]; ok {
		return cached
	}

	engine := getNativeEngineByFS(t)
	if engine == nil {
		engine = reg
	}
	fsCache[t] = engine
	return engine
}

func (f *FileStore) Get(name string) ([]byte, error) {
	if f.engine == nil {
		return nil, os.ErrNotExist
	}
	return f.engine.Get(f.path, name)
}

func (f *FileStore) Set(name string, value []byte) error {
	if f.engine == nil {
		return errors.New("no storage engine available")
	}
	return f.engine.Set(f.path, name, value)
}

func (f *FileStore) List() ([]string, error) {
	if f.engine == nil {
		return nil, nil
	}
	return f.engine.List(f.path)
}

func (f *FileStore) Remove(name string) error {
	if f.engine == nil {
		return nil
	}
	return f.engine.Remove(f.path, name)
}
