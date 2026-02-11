// Copyright (c) 2026 GNOMATIX. All rights reserved.

package xyzattr_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"gitea.gnomatix.com/gnomatix/xyzattr/pkg/xyzattr"
)

func TestStorageIntegrity(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "integrity_test.txt")
	if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	store := xyzattr.At(path)

	// Test data
	attrs := map[string][]byte{
		"gnomatix.owner": []byte("killbot"),
		"gnomatix.mode":  []byte("active"),
		"gnomatix.id":    {0xDE, 0xAD, 0xBE, 0xEF},
	}

	// The library MUST NOT fail. It always falls back if native FS support is missing.
	for name, val := range attrs {
		if err := store.Set(name, val); err != nil {
			t.Fatalf("Set failed for %s: %v (Violates 'Never-Fails' requirement)", name, err)
		}
	}

	// Verify persistence and correctness
	for name, expected := range attrs {
		actual, err := store.Get(name)
		if err != nil {
			t.Fatalf("Get failed for %s: %v", name, err)
		}
		if !bytes.Equal(actual, expected) {
			t.Errorf("Attribute %s: expected %v, got %v", name, expected, actual)
		}
	}

	// Verify listing
	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	for name := range attrs {
		found := false
		for _, n := range list {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Attribute %s missing from list %v", name, list)
		}
	}

	// Verify removal
	target := "gnomatix.mode"
	if err := store.Remove(target); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	_, err = store.Get(target)
	if err == nil {
		t.Errorf("Attribute %s should be removed, but Get succeeded", target)
	}
}
