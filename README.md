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

    "github.com/unstoppablemango/ihfs/osfs"
    "github.com/unstoppablemango/ihfs"
)

var fs fs.FS = osfs.New()

if mkdir, ok := fs.(ihfs.Mkdir); ok {
    _ = mkdir.Mkdir("foo", os.ModeDir)
}

if w, ok := fs.(ihfs.WriteFile); ok {
    _, _ = w.WriteFile("foo/bar.txt", []byte("❤️"), os.ModePerm)
}

seq, err := ihfs.Catch(ihfs.Iter(fs, "."))

for path, dirEntry := range seq {
    // .
    // ./foo
    // ./foo/bar.txt
}
```
