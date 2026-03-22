# osfs

Thin wrapper around `github.com/unmango/go/os`, adapting it to the `ihfs.FS` interface.

## Design

- All operations delegate directly to the external `os` package wrapper
- No business logic lives here; this is a pure adapter
- `osfs.Default` provides a package-level default instance

## Coverage

Coverage not required — this package contains no business logic.
