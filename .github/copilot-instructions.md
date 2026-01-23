# Agent Instructions for IHFS

## Project Overview

IHFS (I ❤️ File Systems) is a Go library providing composable filesystem interfaces, similar to afero but more aligned with Go's `io/fs` package philosophy.

## Technology Stack

- **Language**: Go 1.25.5
- **Testing**: Ginkgo v2 + Gomega
- **Build System**: Nix + Make
- **Package Manager**: Go modules with gomod2nix

## Building and Testing

### Build Commands
```bash
# Build with Nix
nix build .#ihfs

# Run tests
make test
# OR
go tool ginkgo -r

# Generate coverage
make cover

# Format code
make fmt
# OR
nix fmt
```

### Testing Guidelines
- Use Ginkgo/Gomega for all tests
- Test files follow `*_test.go` convention
- Suite tests use `*_suite_test.go` pattern
- Run tests recursively with `ginkgo -r`
- Maintain test coverage (check with `make cover`)

## Code Conventions

### Go Style
- Follow standard Go formatting (gofmt)
- Use tabs for indentation (Go default)
- Insert final newlines in all files
- Trim trailing whitespace
- Keep interfaces small and composable
- Use type aliases for standard library types when appropriate

### Package Structure
- Type aliases in `fs.go` for standard interfaces
- Custom operations defined in `op.go`
- Implementation packages in subdirectories (e.g., `osfs/`, `testfs/`)
- Iterator utilities in `iter.go`

### Interface Design
- Prefer composable, single-purpose interfaces
- Follow `io/fs` patterns and conventions
- Use interface compliance checks: `var _ Interface = (*Type)(nil)`
- Type check with `ok` idiom before calling interface methods

### Naming Conventions
- Use standard Go naming conventions
- FS-related types use abbreviated names (e.g., `FS`, not `FileSystem`)
- Public APIs should be clear and concise
- Avoid stuttering (e.g., `ihfs.FS` not `ihfs.IHFSFS`)

## Dependencies
- **Core**: Standard library `io/fs` package
- **External**: `github.com/unmango/go/os` for OS filesystem
- **Testing**: Ginkgo v2 and Gomega
- **Tools**: gomod2nix for Nix integration

## Development Workflow

1. Make changes to Go source files
2. Run `make test` to ensure tests pass
3. Run `make fmt` to format code
4. Check coverage with `make cover` if modifying core logic
5. Update `go.mod` if adding dependencies, then run `go tool gomod2nix`

## Important Notes

- This project uses Nix for reproducible builds
- EditorConfig settings should be respected
- All public APIs should have clear documentation
- Keep the library focused on composable filesystem abstractions
- Maintain compatibility with Go's `io/fs` package philosophy
- Test data goes in `testdata/` directory

## Common Tasks

### Adding a New Filesystem Type
1. Create a new package directory (e.g., `newfs/`)
2. Implement required `fs.FS` interface
3. Add additional interfaces as needed (ReadDir, Stat, etc.)
4. Add interface compliance checks
5. Write comprehensive tests using Ginkgo

### Adding a New Operation
1. Define operation interface/type in `op.go` or `op/` package
2. Ensure it implements `Operation` interface (has `Path()` method)
3. Add tests for the operation
4. Document the operation's purpose and usage

### Modifying Core Interfaces
1. Check for breaking changes to public API
2. Update all implementations
3. Update tests across all packages
4. Verify with `make test` and `make cover`
