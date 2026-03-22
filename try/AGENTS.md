# try

Type-safe wrappers that attempt optional filesystem interface methods.

## Design

- Each function accepts an `ihfs.FS` or `ihfs.File`, checks whether it implements an optional interface, and calls it if so
- Returns an error wrapping `ihfs.ErrNotImplemented` (detectable with `errors.Is`) when the interface is not satisfied — never panics
- No side effects; purely functional

## Coverage

100% coverage required. All branches (interface supported / not supported) must be tested.
