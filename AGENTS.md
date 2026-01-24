# AI Agent Instructions for IHFS

This document provides guidance for AI agents working with the IHFS (I ❤️ File Systems) codebase.

## Project Overview

IHFS is a Go library providing composable filesystem interfaces, similar to afero but more aligned with Go's `io/fs` package philosophy. The library focuses on small, composable interfaces that can be combined to build complex filesystem abstractions.

## Technology Stack

- **Language**: Go (see go.mod for version)
- **Testing Framework**: Ginkgo v2 + Gomega
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
- Test data goes in `testdata/` directory

## Code Conventions

### Go Style

- Follow standard Go formatting (gofmt)
- Use tabs for indentation (Go default)
- Insert final newlines in all files
- Trim trailing whitespace
- Keep interfaces small and composable
- Use type aliases for standard library types when appropriate
- Only comment code that needs clarification; avoid obvious comments

### Package Structure

- Type aliases in `fs.go` for standard interfaces
- Custom operations defined in `op.go`
- Implementation packages in subdirectories (e.g., `osfs/`, `testfs/`)
- Iterator utilities in `iter.go`
- Utility functions in `fsutil/` package

### Interface Design

- Prefer composable, single-purpose interfaces
- Follow `io/fs` patterns and conventions
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

## Project Structure

```
.
├── fs.go              # Type aliases for standard fs interfaces
├── iter.go            # Iterator utilities for filesystem traversal
├── op.go              # Operation interface definitions
├── op/                # Operation implementations
├── fsutil/            # Filesystem utility functions
│   ├── util.go        # Utility functions for Stat interface
│   └── try/           # Type-safe utility functions with interface checks
├── osfs/              # OS filesystem implementation
├── testfs/            # Test filesystem utilities
└── testdata/          # Test data files
```

## Common Tasks

### Adding a New Filesystem Type

1. Create a new package directory (e.g., `newfs/`)
2. Implement required `fs.FS` interface
3. Add additional interfaces as needed (ReadDir, Stat, etc.)
4. Write comprehensive tests using Ginkgo

Example:
```go
type MyFS struct {
    // fields
}

func (fs *MyFS) Open(name string) (ihfs.File, error) {
    // implementation
}
```

### Adding a New Operation

1. Define operation interface/type in `op/` package
2. Ensure it implements `Operation` interface (has `Subject()` method)
3. Add tests for the operation
4. Document the operation's purpose and usage

### Modifying Core Interfaces

1. Check for breaking changes to public API
2. Update all implementations
3. Update tests across all packages
4. Verify with `make test` and `make cover`

### Writing Tests

Use Ginkgo's BDD-style testing:

```go
var _ = Describe("MyFeature", func() {
    It("should do something", func() {
        result := MyFunction()
        Expect(result).To(BeTrue())
    })
})
```

For mocking filesystems, use the `testfs` package or create simple mock implementations.

## Self-Correction

When working with this codebase, agents should self-correct and improve documentation:

- **If the code map is discovered to be stale, update it.** The project structure section and other documentation should reflect the current state of the repository. When you discover inaccuracies, update this document.

- **If the user gives a correction about how work should be done in this repo, add it to "Local Norms" (or another clearly labeled section) so future sessions inherit it.** User feedback about repository-specific practices should be captured for future reference.

## Local Norms

This section contains repository-specific practices learned from user feedback:

- Test coverage should be maintained at 100% when possible
- Use mock implementations in tests rather than complex test fixtures

## Important Notes

- This project uses Nix for reproducible builds
- EditorConfig settings should be respected
- All public APIs should have clear documentation
- Keep the library focused on composable filesystem abstractions
- Maintain compatibility with Go's `io/fs` package philosophy
- Minimize dependencies to keep the library lightweight

## File Naming Conventions

- Source files: `feature.go`
- Test files: `feature_test.go`
- Test suites: `package_suite_test.go`
- Internal packages: `internal/package/`

## Error Handling

- Use standard `io/fs` errors when applicable (`fs.ErrNotExist`, `fs.ErrPermission`, etc.)
- Wrap errors with context using `fmt.Errorf` with `%w` verb
- Return clear, actionable error messages

## Performance Considerations

- Avoid unnecessary allocations
- Use buffers and pools where appropriate
- Consider lazy evaluation for expensive operations
- Profile performance-critical code paths

## Documentation

- All exported types, functions, and methods must have documentation comments
- Use Go's standard documentation format
- Include examples in documentation where helpful
- Keep documentation up-to-date with code changes
