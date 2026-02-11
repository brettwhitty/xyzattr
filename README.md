# xyzattr

## Project Description
Simple Golang library for performant, stable, persistent, cross-platform, cross-filesystem abstraction of extensible, schema-driven extended file attributes.

## Development Notes

### Language

- Golang (1.25)

### Requirements

#### Cross-Platform

Must compile and provide consistent function on at least:

- Windows
- macOS
- Linux

#### Cross-Filesystem

The core functionality of this library is to provide all features and functions of POSIX-compliant `xattr` by abstraction and/or virtualization.

##### Native `xattr` Support

On systems / fileysystems with native POSIX-compliant `xattr` support, this library should provide native-level functionality.

See: https://en.wikipedia.org/wiki/Extended_file_attributes

##### Alternative Native Persistent Attribute Store

Where native `xattr` support is lacking, alternative OS-/filesystem-specific extended attribute methods should be supported and configurable.

For example:
- ZFS properties
- NTFS alternative data streams (see `NTFS-3G` implementation)

_NOTE_ Supported alternative storage methods must be:
- Most sensible/appropriate for the particular OS/filesystem.
- Non-destructive --- should not alter the file itself.
- Minimal-to-Zero impact
- Performant

##### Alternative Standard/Libary-specific Persistent Attribute Store

The library should be implemented in such a way that failing native `xattr`, and native `xattr`-like support, it will fall back on context-appropriate alternatives.

The fallback strategy should appropriately choose the best implementation based on:
- OS
- Filesystem
- Performance
- Reliability
- Standards-compliant
- Minimal-to-Zero impact

### Essential Library Abstractions

All POSIX-compliant `xattr` functionality should be abstracted at the top-level of implementation.

#### Example: Polymorphic Get / Set
Simple functions like get/set might be abstracted as:

```
att := xyzattr('path/to/file').Get('att')
att := xyzattr('path/to/file').Set('att', 'val')
```

Where the 'Get' function automatically determines/verifies the best-available datastore for the particular file.

- The behavior should be consistent and safe.
  Eg: it should determine and cache "best" storage for each filesystem on first access, not with each operation.
- It should require configuration where no obvious "best" implementation exists.

## ESSENTIAL REQUIREMENT

### NON-DESTRUCTIVE

The library must not in any way modify the file's contents themselves.

- No implemenetations can be considered that would in any way alter the original file.

## Usability

### As a Library

First and foremost, this must function as a library available for use in the development of other cross-platform Golang software.

- Full, abstracted POSIX-compliant `xattr` functionality is therefore ideal.
- Clients of the library should not have to care what is going on "under the hood".

However, it is worth considering while developing the abstraction layers that:

- Sensible fallback storage layers (eg: database?) might be client-configured / provided.
- We may want to provide / suggest a fallback storage layer and make it available to the client.

### CLI Tools

We should ultimately provide command-line tools that provide the same POSIX-compliant `xattr` functionality that is available on OSes like Linux, Solaris and macOS.

Again, unlike the `xattr` tools, these tools _must_ be runnable, and provide the same functionality on any filesystem on the host OS.

## Sidecar Files

### AVOID SIDECAR FILES

Let's not use sidecar files if we can avoid it.

