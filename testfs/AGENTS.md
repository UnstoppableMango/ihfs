# testfs

Configurable mock filesystem for use in tests across the repo.

## Key Types

- `Fs`: mock filesystem; each operation has a corresponding `*Func` field (e.g., `OpenFunc`, `StatFunc`)
- `File`: mock file with per-operation function fields
- `FileInfo`: mock `fs.FileInfo`; create with `testfs.NewFileInfo(name)`
- `factory.Fs`: queue-based factory — returns a different mock per call, useful for testing sequences

## Usage Patterns

```go
// Simple mock
fsys := testfs.New(testfs.WithOpen(func(name string) (ihfs.File, error) {
    return nil, ihfs.ErrNotExist
}))

// FileInfo
info := testfs.NewFileInfo("example.txt")
```

## Constraints

- There is no `testfs.BoringFileInfo` — use `testfs.NewFileInfo()` or `testfs.FileInfo` directly
- `factory.Fs` dequeues one mock per call; if the queue is empty it returns an error — always enqueue enough mocks for the test

## Coverage

Coverage not required — this is a test helper package.
