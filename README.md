# I ❤️ File Systems

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/UnstoppableMango/ihfs/ci.yml)
![GitHub branch check runs](https://img.shields.io/github/check-runs/UnstoppableMango/ihfs/main)
![Codecov](https://img.shields.io/codecov/c/github/UnstoppableMango/ihfs)

Similar to [afero](https://github.com/spf13/afero), but with composable interfaces more akin to `io/fs`.

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
```

## GitHub FS

Package `ghfs` contains an implementation of `io/fs` for the GitHub API.

[Documentation](./ghfs/README.md)

## Attribution

Much of the implementation is adapted from [afero](https://github.com/spf13/afero), specifically the `corfs`, `cowfs`, and `union` packages.

> Imitation is the sincerest form of flattery

- [Charles Caleb Colton](https://shkspr.mobi/blog/2024/01/no-oscar-wilde-did-not-say-imitation-is-the-sincerest-form-of-flattery-that-mediocrity-can-pay-to-greatness/)
