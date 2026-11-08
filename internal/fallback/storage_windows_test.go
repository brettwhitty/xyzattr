// Copyright (c) 2026 GNOMATIX. All rights reserved.

package fallback

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestRegistryReplacementFailurePreservesData(t *testing.T) {
	for _, operation := range []string{"set", "remove"} {
		t.Run(operation, func(t *testing.T) {
			path := fixtureHome(t)
			s := openStore(t)
			if err := s.Set("p", "a", []byte("saved")); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			// Go's regular read handle permits read/write sharing, but not deletion.
			// It allows staging a replacement and makes the final rename fail.
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if operation == "set" {
				err = s.Set("p", "a", []byte("failed"))
			} else {
				err = s.Remove("p", "a")
			}
			var linkErr *os.LinkError
			if !errors.As(err, &linkErr) || linkErr.Op != "rename" || (!errors.Is(err, syscall.ERROR_ACCESS_DENIED) && !errors.Is(err, syscall.Errno(32))) {
				t.Fatalf("error = %v; want native replacement denial", err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("committed bytes changed: %q, %v", after, err)
			}
			requireValue(t, s, "p", "a", []byte("saved"))
			requireValue(t, openStore(t), "p", "a", []byte("saved"))
			files, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".attributes-*"))
			if err != nil || len(files) != 0 {
				t.Fatalf("staging files left after failure: %v, %v", files, err)
			}
		})
	}
}
