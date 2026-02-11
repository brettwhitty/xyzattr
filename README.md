# xyzattr

![xyzattr](./assets/images/xyzattr-banner-lets-go.jpg "xyzattr --- Let's GO!!!")

Simple Golang library for performant, stable, persistent, cross-platform, cross-filesystem abstraction of extensible, schema-driven extended file attributes.

## Features

- **Cross-Platform**: Consistent API for Windows, macOS, and Linux.
- **Cross-Filesystem**: Support for native `xattr`, NTFS Alternate Data Streams, ZFS properties, and smart fallback mechanisms.
- **Non-Destructive**: Never modifies the original file contents.
- **Performant**: Designed for minimal overhead and stable persistence.

## Documentation

Comprehensive details on implementation and requirements can be found in the [Technical Specifications](docs/spec.md).

## Usage

The `xyzattr` library provides a simple, fluent API that guarantees persistence regardless of the underlying filesystem.

```go
import "gitea.gnomatix.com/gnomatix/xyzattr/pkg/xyzattr"

func main() {
    // Get a store bound to a file
    store := xyzattr.At("path/to/file.txt")

    // Set an attribute (automatically uses native ADS/xattr or falls back to registry)
    err := store.Set("user.comment", []byte("Important file"))

    // Get an attribute
    val, err := store.Get("user.comment")

    // List all attributes
    list, err := store.List()
}
```

## Development

This project uses [mise](https://mise.jdx.co/) for development tasks.

```bash
# Build the project
mise run build

# Run tests
mise run test

# Tidy Go modules
mise run tidy
```

## License

BSD 3-Clause License

Copyright (c) 2026, GNOMATIX
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,   
   this list of conditions and the following disclaimer in the documentation and/or
   other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its
   contributors may be used to endorse or promote products derived from
   this software without specific prior written permission.

---

![Killbots Activate!](./assets/images/gnomatix-killbots-activate-xs.png "Go Team!")
![GNOMATIX](./assets/images/gnomatix-new-xs.png "GNOMATIX")