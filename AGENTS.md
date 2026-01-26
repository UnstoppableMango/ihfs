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

- Type aliases and error constants in `fs.go` for standard interfaces
- `Operation` interface defined in `fs.go`
- Concrete operation types in `op/` package
- Implementation packages in subdirectories (e.g., `osfs/`, `cowfs/`, `tarfs/`, `testfs/`)
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

## Codebase Map

### Module Information
- **Module Path**: `github.com/unstoppablemango/ihfs`
- **Go Version**: 1.25.5

### Entry Points & Core Files

#### Root Package (`./`)
- **`fs.go`**: Type aliases for `io/fs` interfaces + standard error aliases + `Operation` interface definition + custom FS interfaces (Chmod, Chown, Chtimes, Copy, Mkdir, etc.)
- **`file.go`**: Type aliases for file-related interfaces (`File`, `FileInfo`, `DirEntry`, `FileMode`, `PathError`) + standard error aliases + `Operation` interface definition + `Seeker` interface
- **`iter.go`**: Iterator utilities for traversing filesystems (`Iter`, `Catch` functions)

#### Implementation Packages
- **`osfs/fs.go`**: OS filesystem implementation (wraps `github.com/unmango/go/os`)
- **`cowfs/`**: Copy-on-write filesystem implementation
  - `fs.go`: Copy-on-write filesystem (base + overlay layers)
  - `file.go`: Copy-on-write file implementation
  - `bsds.go`: BSD-specific constants (EBADF)
  - `win_unix.go`: Windows/Unix-specific constants (EBADFD)
  - `doc.go`: Package documentation
- **`testfs/`**: Test filesystem utilities
  - `fs.go`: Test filesystem implementation
  - `file.go`: Test file implementation
  - `fileinfo.go`: Test FileInfo implementation
  - `option.go`: Option pattern for test setup
- **`tarfs/`**: Tar filesystem implementation
  - `fs.go`: Tar filesystem implementation
  - `file.go`: Tar file implementation
  - `cache.go`: Caching utilities for tar entries
  - `doc.go`: Package documentation

##### Filesystem Implementation Overview
- **osfs**: Wraps the OS filesystem for standard file operations
- **cowfs**: Copy-on-write filesystem with base and overlay layers (based on afero.CopyOnWriteFs)
  - Changes only affect the overlay layer
  - Reads prioritize overlay over base
  - Directories from both layers are merged
  - Constructor: `cowfs.New(base, layer ihfs.FS) *Fs`
- **tarfs**: Read-only filesystem backed by tar archives
- **testfs**: Mock filesystem for testing with configurable behavior

#### Operation Types
- **`op/`**: File system operation definitions
  - `doc.go`: Package documentation
  - `operation.go`: Concrete operation type implementations

#### Utilities
- **`fsutil/fs.go`**: Filesystem utilities (FS-related helpers)
- **`fsutil/try/`**: Error-handling utilities for FS operations
  - `fs.go`: Type-safe wrappers for FS operations with interface checks
  - `file.go`: Type-safe wrappers for File operations with interface checks

### Testing Structure

#### Test Framework
- **Ginkgo v2 + Gomega** for all tests
- Run: `make test` or `go tool ginkgo -r`

#### Test Files by Package
- **Root (`ihfs_test`)**: `ihfs_suite_test.go`, `iter_test.go`
- **fsutil (`fsutil_test`)**: `fsutil_suite_test.go`, `fs_test.go`
- **fsutil/try (`try_test`)**: `try_suite_test.go`, `fs_test.go`, `file_test.go`
- **cowfs (`cowfs_test`)**: `cowfs_suite_test.go`, `fs_test.go`, `file_test.go`
- **tarfs (`tarfs_test`)**: `tarfs_suite_test.go`, `fs_test.go`, `file_test.go`

#### Test Data
- **`testdata/`**: Test fixtures and sample files
  - `2-files/`: Fixture with two files for testing
  - `test.tar`: Tar archive for testing tar filesystem

### Build & CI Configuration
- **`Makefile`**: Build targets (`build`, `test`, `cover`, `fmt`)
- **`.github/workflows/ci.yml`**: GitHub Actions CI pipeline
- **`flake.nix`**: Nix build and development environment
- **`gomod2nix.toml`**: Nix-Go module integration

### Package Naming Conventions
- **Main package**: `ihfs` (core library code)
- **Tests**: `ihfs_test`, `fsutil_test`, `try_test`, `cowfs_test` (external test packages)
- **Implementations**: Named after their purpose (`osfs`, `cowfs`, `tarfs`, `testfs`)
- **Test suites**: Follow `*_suite_test.go` pattern
- **Test files**: Follow `*_test.go` pattern

### Key Dependencies
- **`io/fs`** (stdlib): Base filesystem interfaces
- **`github.com/unmango/go/os`**: OS filesystem wrapper
- **Ginkgo v2 / Gomega**: Testing framework

### Project Structure

```
.
├── fs.go              # Type aliases, error aliases, Operation interface, and custom FS interfaces
├── file.go            # File-related type aliases and Seeker interface
├── iter.go            # Iterator utilities for filesystem traversal
├── op/                # Concrete operation type implementations
│   ├── doc.go         # Package documentation
│   └── operation.go   # Operation implementations
├── fsutil/            # Filesystem utility functions
│   ├── fs.go          # FS-related utilities
│   └── try/           # Type-safe utility functions with interface checks
│       ├── fs.go      # FS operation wrappers
│       └── file.go    # File operation wrappers
├── osfs/              # OS filesystem implementation
│   └── fs.go          # OS filesystem wrapper
├── cowfs/             # Copy-on-write filesystem implementation
│   ├── fs.go          # Copy-on-write filesystem (base + overlay layers)
│   ├── file.go        # Copy-on-write file implementation
│   ├── bsds.go        # BSD-specific constants
│   ├── win_unix.go    # Windows/Unix-specific constants
│   └── doc.go         # Package documentation
├── tarfs/             # Tar filesystem implementation
│   ├── fs.go          # Tar filesystem
│   ├── file.go        # Tar file implementation
│   ├── cache.go       # Caching utilities
│   └── doc.go         # Package documentation
├── testfs/            # Test filesystem utilities
│   ├── fs.go          # Test filesystem
│   ├── file.go        # Test file implementation
│   ├── fileinfo.go    # Test FileInfo
│   └── option.go      # Test setup options
└── testdata/          # Test data files
    ├── 2-files/       # Test fixture with two files
    └── test.tar       # Tar archive for testing
```

## Common Tasks

### Adding a New Filesystem Type

1. Create a new package directory (e.g., `newfs/`)
2. Implement required `fs.FS` interface
3. Add additional interfaces as needed (ReadDir, Stat, etc.)
4. Write comprehensive tests using Ginkgo

Example:
```go
type Fs struct {
    // fields
}

func (fs *Fs) Open(name string) (ihfs.File, error) {
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
