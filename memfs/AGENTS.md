# memfs

Full-featured in-memory filesystem with read/write support.

## Design

- Root directory is always present and created lazily via `sync.Once`
- Internal node map is keyed by cleaned path
- Thread-safe: all map access is protected by a `sync.RWMutex`
- Supports the full set of filesystem operations: Create, Mkdir, Remove, Rename, Chmod, Chtimes, etc.

## Key Constraints

- Paths are cleaned/normalized before lookup — never store or compare raw user-supplied paths
- Some defensive branches (e.g., empty path components after normalization) are intentionally unreachable; simplify the code rather than writing impossible tests for them

## Coverage

Aim for high coverage. If a branch proves genuinely unreachable, remove it rather than writing impossible tests. Do not write messy or low-value tests just to hit 100%.
