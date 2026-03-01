# Codebase Map

## Module Information

- **Module Path**: `github.com/unstoppablemango/ihfs`
- **Go Version**: Refer to go.mod

## Entry Points & Core Files

### Root Package (`./`)

- **`fs.go`**: Type aliases for `io/fs` interfaces + standard error aliases + `Operation` interface definition + custom FS interfaces (Chmod, Chown, Chtimes, Copy, Mkdir, etc.)
- **`file.go`**: Type aliases for file-related interfaces (`File`, `FileInfo`, `DirEntry`, `FileMode`, `PathError`) + standard error aliases + `Operation` interface definition + `Seeker` interface
- **`iter.go`**: Iterator utilities for traversing filesystems (`Iter`, `Catch` functions)

### Implementation Packages

- **`osfs/fs.go`**: OS filesystem implementation (wraps `github.com/unmango/go/os`)
- **`cowfs/`**: Copy-on-write filesystem implementation
  - `fs.go`: Copy-on-write filesystem (base + layer)
  - `option.go`: Configuration options
  - `doc.go`: Package documentation
- **`corfs/`**: Cache-on-read filesystem implementation (based on afero.CacheOnReadFs)
  - `fs.go`: Cache-on-read filesystem (base + layer with caching)
  - `option.go`: Configuration options (cache time)
  - `doc.go`: Package documentation
- **`union/`**: Union filesystem utilities
  - `copy.go`: File copying utilities for layered filesystems
  - `file.go`: Union file implementation (merges base and layer files)
  - `merge.go`: Directory entry merging strategies
  - `option.go`: Configuration options (merge strategy)
  - `bsds.go`: BSD-specific constants (BADFD)
  - `win_unix.go`: Windows/Unix-specific constants (BADFD)
- **`testfs/`**: Test filesystem utilities
  - `fs.go`: Test filesystem implementation
  - `file.go`: Test file implementation
  - `fileinfo.go`: Test FileInfo implementation
  - `option.go`: Option pattern for test setup
  - `boring.go`: Boring implementation helpers
  - `testfs.go`: Additional test utilities
  - `factory/fs.go`: Queue-based factory filesystem for per-call mock control
- **`tarfs/`**: Tar filesystem implementation
  - `fs.go`: Tar filesystem implementation
  - `file.go`: Tar file implementation
  - `cache.go`: Caching utilities for tar entries
  - `doc.go`: Package documentation
- **`memfs/`**: In-memory filesystem implementation
  - `fs.go`: In-memory filesystem implementation with full read/write support
  - `file.go`: In-memory file implementation with read/write capabilities
  - `fileinfo.go`: FileInfo implementation for in-memory files

### Filesystem Implementation Overview

- **osfs**: Wraps the OS filesystem for standard file operations
  - Simple wrapper around `github.com/unmango/go/os`
- **cowfs**: Copy-on-write filesystem with base and layer (based on afero.CopyOnWriteFs)
  - Changes only affect the layer
  - Reads prioritize layer over base
  - Directories from both layers are merged
  - Constructor: `cowfs.New(base, layer ihfs.FS, options ...union.Option) *Fs`
- **corfs**: Cache-on-read filesystem (based on afero.CacheOnReadFs)
  - Files are cached from base to layer on first read
  - Future reads come from cached version
  - Configurable cache expiration time
  - Constructor: `corfs.New(base, layer ihfs.FS, options ...Option) *Fs`
- **union**: Utilities for union/layered filesystems
  - `CopyToLayer`: Copies files from base to layer with metadata preservation
  - `NewFile`: Creates union file that merges base and layer file operations
  - `mergeDirEntries`: Strategies for merging directory entries from multiple layers
- **tarfs**: Read-only filesystem backed by tar archives
- **memfs**: Full-featured in-memory filesystem implementation
  - Complete read/write support for files and directories
  - Thread-safe operations with mutex locking
  - Supports standard filesystem operations (Create, Mkdir, Remove, Rename, Chmod, etc.)
  - Constructor: `memfs.New() *Fs`
- **testfs**: Mock filesystem for testing with configurable behavior

### Operation Types

- **`op/`**: File system operation definitions
  - `doc.go`: Package documentation
  - `operation.go`: Concrete operation type implementations

### Utilities

- **`try/`**: Error-handling utilities for FS operations
  - `fs.go`: Type-safe wrappers for FS operations with interface checks
  - `file.go`: Type-safe wrappers for File operations with interface checks

## Testing Structure

### Test Framework

- **Ginkgo v2 + Gomega** for all tests
- Run: `make test` or `go tool ginkgo -r`

### Test Files by Package

- **Root (`ihfs_test`)**: `ihfs_suite_test.go`, `iter_test.go`, `filter_test.go`, `util_test.go`
- **try (`try_test`)**: `try_suite_test.go`, `fs_test.go`, `file_test.go`
- **cowfs (`cowfs_test`)**: `cowfs_suite_test.go`, `fs_test.go`
- **corfs (`corfs_test`)**: `corfs_suite_test.go`, `fs_test.go`
- **union (`union_test`)**: `union_suite_test.go`, `copy_test.go`, `file_test.go`, `merge_test.go`
- **tarfs (`tarfs_test`)**: `tarfs_suite_test.go`, `fs_test.go`, `file_test.go`
- **memfs (`memfs_test`)**: `memfs_suite_test.go`, `fs_test.go`

### Test Data

- **`testdata/`**: Test fixtures and sample files
  - `2-files/`: Fixture with two files for testing
  - `test.tar`: Tar archive for testing tar filesystem

## Build & CI Configuration

- **`Makefile`**: Build targets (`build`, `test`, `cover`, `fmt`)
- **`.github/workflows/ci.yml`**: GitHub Actions CI pipeline
- **`flake.nix`**: Nix build and development environment
- **`gomod2nix.toml`**: Nix-Go module integration

## Package Naming Conventions

- **Main package**: `ihfs` (core library code)
- **Tests**: `ihfs_test`, `try_test`, `cowfs_test`, `corfs_test`, `union_test`, `tarfs_test`, `memfs_test` (external test packages)
- **Implementations**: Named after their purpose (`osfs`, `cowfs`, `corfs`, `tarfs`, `memfs`, `testfs`)
- **Utilities**: `union` for layered filesystem utilities
- **Test suites**: Follow `*_suite_test.go` pattern
- **Test files**: Follow `*_test.go` pattern

## Key Dependencies

- **`io/fs`** (stdlib): Base filesystem interfaces
- **`github.com/unmango/go/os`**: OS filesystem wrapper
- **Ginkgo v2 / Gomega**: Testing framework

## Project Structure

```
.
├── fs.go              # Type aliases, error aliases, Operation interface, and custom FS interfaces
├── file.go            # File-related type aliases and Seeker interface
├── iter.go            # Iterator utilities for filesystem traversal
├── op/                # Concrete operation type implementations
│   ├── doc.go         # Package documentation
│   └── operation.go   # Operation implementations
├── try/               # Type-safe utility functions with interface checks
│   ├── fs.go          # FS operation wrappers
│   └── file.go        # File operation wrappers
├── osfs/              # OS filesystem implementation
│   └── fs.go          # OS filesystem wrapper
├── cowfs/             # Copy-on-write filesystem implementation
│   ├── fs.go          # Copy-on-write filesystem (base + layer)
│   ├── option.go      # Configuration options
│   └── doc.go         # Package documentation
├── corfs/             # Cache-on-read filesystem implementation
│   ├── fs.go          # Cache-on-read filesystem (base + layer with caching)
│   ├── option.go      # Configuration options (cache time)
│   └── doc.go         # Package documentation
├── union/             # Union filesystem utilities
│   ├── copy.go        # File copying utilities for layered filesystems
│   ├── file.go        # Union file implementation (merges base and layer files)
│   ├── merge.go       # Directory entry merging strategies
│   ├── option.go      # Configuration options (merge strategy)
│   ├── bsds.go        # BSD-specific constants (BADFD)
│   └── win_unix.go    # Windows/Unix-specific constants (BADFD)
├── tarfs/             # Tar filesystem implementation
│   ├── fs.go          # Tar filesystem
│   ├── file.go        # Tar file implementation
│   ├── cache.go       # Caching utilities
│   └── doc.go         # Package documentation
├── memfs/             # In-memory filesystem implementation
│   ├── fs.go          # In-memory filesystem (thread-safe, full read/write)
│   ├── file.go        # In-memory file implementation
│   └── fileinfo.go    # FileInfo implementation
├── testfs/            # Test filesystem utilities
│   ├── fs.go          # Test filesystem
│   ├── file.go        # Test file implementation
│   ├── fileinfo.go    # Test FileInfo
│   ├── option.go      # Test setup options
│   ├── boring.go      # Boring implementation helpers
│   ├── testfs.go      # Additional test utilities
│   └── factory/
│       └── fs.go      # Queue-based factory filesystem for per-call mock control
└── testdata/          # Test data files
    ├── 2-files/       # Test fixture with two files
    └── test.tar       # Tar archive for testing tar filesystem
```
