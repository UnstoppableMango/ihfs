# union

Core abstractions for layered (copy-on-write / cache-on-read) filesystems.

## Key Types

- `File`: wraps a base file and a layer file; reads prefer layer, writes go to both
- `MergeStrategy`: function type controlling how directory entries from two layers are combined
- `CopyToLayer`: copies a file from base to layer, preserving mtime

## Design Constraints

- `File.Read()` reads from layer, then syncs the offset to base (so both stay in sync)
- `File.Write()` writes to layer first, then base
- `ReadDir()` merges entries lazily using the configured `MergeStrategy`
- `CopyToLayer` must create all parent directories before writing the file
- Always check `io.Writer` interface before attempting writes to a file handle

## Relationships

- Used by `cowfs` and `corfs` — do not introduce cowfs/corfs-specific logic here
- Options are passed through from cowfs/corfs constructors

## Coverage

100% coverage required.
