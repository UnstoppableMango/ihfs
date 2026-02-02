# I ❤️ File Systems

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/UnstoppableMango/ihfs/ci.yml)
![GitHub branch check runs](https://img.shields.io/github/check-runs/UnstoppableMango/ihfs/main)
![Codecov](https://img.shields.io/codecov/c/github/UnstoppableMango/ihfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/unstoppablemango/ihfs)](https://goreportcard.com/report/github.com/unstoppablemango/ihfs)
![Go version](https://img.shields.io/github/go-mod/go-version/UnstoppableMango/ihfs)

Similar to [afero](https://github.com/spf13/afero), but built around the extension interface pattern described in the [io/fs draft design](https://github.com/golang/proposal/blob/master/design/draft-iofs.md).

Also an anagram for "fish".

This was not intentional.

## Usage

```go
import (
    "os"
    "io/fs"

    "github.com/unstoppablemango/ihfs"
    "github.com/unstoppablemango/ihfs/osfs"
    "github.com/unstoppablemango/ihfs/try"
)

// Built around [io/fs]
var fs fs.FS = osfs.New()

// Regular type checks
if mkdir, ok := fs.(ihfs.MkdirFS); ok {
    _ = mkdir.Mkdir("foo", os.ModeDir)
}

// The [try] package
_, err := try.WriteFile(fs, "foo/bar.txt", []byte("❤️"), os.ModePerm)

// Walking with [iter.Seq]
seq, err := ihfs.Catch(ihfs.Iter(fs, "."))

for path, dirEntry := range seq {
    // .
    // ./foo
    // ./foo/bar.txt
}

// Filtering
filtered := ihfs.Where(fs, func(o ihfs.Operation) bool {
    return strings.Contains(o.Subject(), "bar.txt")
})

for path, err := range ihfs.IterPaths(filtered) {
    // ./foo/bar.txt
}
```

## Implementations

### osfs

Wraps the OS filesystem. The `Default` variable provides a package-level instance.

```go
import "github.com/unstoppablemango/ihfs/osfs"

fs := osfs.New()
f, err := fs.Open("path/to/file")
```

### memfs

A full-featured in-memory filesystem with read/write support.
Useful for testing or ephemeral scratch space.

```go
import "github.com/unstoppablemango/ihfs/memfs"

fs := memfs.New()
f, _ := fs.Create("hello.txt")
f.Write([]byte("hello"))
f.Close()
```

### tarfs

A read-only filesystem backed by a tar archive.
Entries are lazily buffered as files are accessed.

```go
import "github.com/unstoppablemango/ihfs/tarfs"

tfs, err := tarfs.Open("archive.tar")
defer tfs.Close()

f, err := tfs.Open("dir/file.txt")
```

You can also construct one from any `io.Reader`:

```go
tfs := tarfs.FromReader("archive.tar", r)
```

### cowfs

A copy-on-write filesystem layered over a base.
All writes go to the layer; reads prefer the layer and fall back to the base.
Modifying a file that exists only in the base copies it to the layer first.

```go
import (
    "github.com/unstoppablemango/ihfs/cowfs"
    "github.com/unstoppablemango/ihfs/memfs"
    "github.com/unstoppablemango/ihfs/osfs"
)

base := osfs.New()
layer := memfs.New()
fs := cowfs.New(base, layer)
```

### corfs

A cache-on-read filesystem.
The first read of a file copies it from the base into the layer; subsequent reads come from the layer.
A cache duration of 0 (the default) caches indefinitely.

```go
import (
    "time"
    "github.com/unstoppablemango/ihfs/corfs"
    "github.com/unstoppablemango/ihfs/memfs"
    "github.com/unstoppablemango/ihfs/osfs"
)

base := osfs.New()
cache := memfs.New()
fs := corfs.New(base, cache)

// With a 5-minute expiry:
fs = corfs.New(base, cache, corfs.WithCacheTime(5*time.Minute))
```

### testfs

Hand-written test doubles with function-field overrides.
Useful for simple unit tests that need a configurable fake filesystem without a full mock framework.

```go
import (
    "github.com/unstoppablemango/ihfs"
    "github.com/unstoppablemango/ihfs/testfs"
)

fs := testfs.New(
    testfs.WithOpen(func(name string) (ihfs.File, error) {
        return testfs.NewFile(name), nil
    }),
)
```

### mockfs

Generated [gomock](https://github.com/uber-go/mock) mocks for all ihfs interfaces.
See the [mockfs README](mockfs/README.md) for details.

## GitHub FS

Package `ghfs` contains an implementation of `io/fs` for the GitHub API.

[Documentation](./ghfs/README.md)

## Attribution

Much of the implementation is adapted from [afero](https://github.com/spf13/afero), specifically the `corfs`, `cowfs`, and `union` packages.
