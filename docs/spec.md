# xyzattr - Technical Specification

This document summarizes the technical requirements, architectural decisions, and internal implementation details for the `xyzattr` library.

## Project Description

Simple Golang library for performant, stable, persistent, cross-platform, cross-filesystem abstraction of extensible, schema-driven extended file attributes.

## Requirements

### Cross-Platform

Must compile and provide consistent function on at least:

- Windows
- macOS
- Linux

### Cross-Filesystem

The core functionality of this library is to provide all features and functions of POSIX-compliant `xattr` by abstraction and/or virtualization.

#### Native `xattr` Support

On systems / fileysystems with native POSIX-compliant `xattr` support, this library should provide native-level functionality.

See: https://en.wikipedia.org/wiki/Extended_file_attributes

#### Alternative Native Persistent Attribute Store

Where native `xattr` support is lacking, alternative OS-/filesystem-specific extended attribute methods should be supported and configurable.

For example:
- ZFS properties
- NTFS alternative data streams (see `NTFS-3G` implementation)

#### Alternative Standard/Library-specific Persistent Attribute Store

The library falls back on context-appropriate alternatives (e.g., database) based on:
- OS
- Filesystem
- Performance
- Reliability
- Standards-compliance
- Minimal-to-Zero impact

### Essential Library Abstractions

All POSIX-compliant `xattr` functionality is abstracted at the top-level of implementation.

#### Polymorphic Get / Set

```go
att := xyzattr("path/to/file").Get("att")
err := xyzattr("path/to/file").Set("att", "val")
```

The `Get` function automatically determines the best-available datastore for the particular file, caching the "best" storage for each filesystem on first access.

## Constraints

### Non-Destructive

The library must not in any way modify the file's contents themselves. No implementations can be considered that would in any way alter the original file.

### Sidecar Files

Avoid sidecar files if at all possible.

---

Copyright (c) 2026 GNOMATIX.
