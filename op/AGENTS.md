# op

Concrete operation types that implement `ihfs.Operation`.

## Design

- Each struct represents one filesystem operation (e.g., `Open`, `Stat`, `ReadDir`)
- Every type implements `Subject() string` returning the primary path or pattern
- Used by `filter/` and `ghfs/` to inspect operations at runtime
- Adding a new interface to the root package should be accompanied by a matching op type here

## Coverage

Coverage not required — these are simple data types with no branching logic.
