# filter

Utilities for `ihfs.FilterFS` — predicate functions that gate filesystem operations.

## Design

- A `FilterFunc` receives the `*ihfs.FilterFS` and the `ihfs.Operation` being attempted, and returns `nil` (allow) or an error (deny)
- `NameRegex(re)` allows operations whose target name matches `re`; directories always pass through (afero parity); ops without a name field (e.g., `op.Glob`) also pass through

## Adding New Filters

1. Add a new function returning `ihfs.FilterFunc`
2. Use a type switch on `ihfs.Operation` to extract the relevant field
3. Document which operation types are handled and which pass through by default

## Relationships

- Depends on `op/` for concrete operation types
- `ihfs.FilterFS` is defined in the root `filter.go`

## Coverage

Aim for high coverage, but do not write messy or low-value tests just to hit 100%. If a branch requires complex setup with little benefit, skip it.
