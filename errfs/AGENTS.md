# errfs

Filesystem and file implementation that return a fixed error for every operation.

## Purpose

Intended exclusively for testing error-handling paths in other packages. Use `errfs.New(err)` to get an `*Fs`, and `errfs.NewFile(err)` to get a `*File`.

## Design

- Both `Fs` and `File` implement wide interfaces (all optional methods) so callers never get `ErrNotImplemented` unexpectedly
- No state beyond the stored error

## Coverage

100% coverage required.
