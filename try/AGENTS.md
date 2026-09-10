# try

Type-safe wrappers that attempt optional filesystem interface methods.

## Design

- Each function accepts an `ihfs.FS` or `ihfs.File`, checks whether it implements an optional interface, and calls it if so
- Returns an error wrapping `ihfs.ErrNotImplemented` (detectable with `errors.Is`) when the interface is not satisfied — never panics
- No extra side effects beyond the delegated filesystem operation

## Coverage

Aim for high coverage. Both branches (interface supported / not supported) should be tested when straightforward, but do not write messy or low-value tests just to hit 100%.
