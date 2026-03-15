# mockfs

[![Go Reference](https://pkg.go.dev/badge/github.com/unstoppablemango/ihfs/mockfs.svg)](https://pkg.go.dev/github.com/unstoppablemango/ihfs/mockfs)
[![Go version](https://img.shields.io/github/go-mod/go-version/UnstoppableMango/ihfs?filename=mockfs/go.mod)](https://github.com/UnstoppableMango/ihfs/blob/main/mockfs/go.mod)

Generated [gomock](https://github.com/uber-go/mock) mocks for all ihfs interfaces.
Provides call tracking, argument matchers, and controller-based verification.

`mockfs` is a separate Go module:

```
go get github.com/unstoppablemango/ihfs/mockfs
```

## Usage

Create a mock, set expectations with `EXPECT()`, and pass it to the code under test.
The gomock controller automatically verifies all expectations at the end of the test.

```go
import (
    "testing"

    "go.uber.org/mock/gomock"
    "github.com/unstoppablemango/ihfs/mockfs"
)

func TestMyCode(t *testing.T) {
    ctrl := gomock.NewController(t)

    fs := mockfs.NewCreateFS(ctrl)
    fs.EXPECT().Create("foo.txt").Return(mockfs.NewFile(ctrl), nil)

    // pass fs to code under test
}
```

### Argument matchers

Use `gomock.Any()` to match any argument, or exact values for strict matching:

```go
fs := mockfs.NewWriteFileFS(ctrl)

// Match any call to WriteFile regardless of arguments
fs.EXPECT().WriteFile(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

// Match only specific arguments
fs.EXPECT().WriteFile("config.json", data, 0o644).Return(nil)
```

### Returning errors

```go
import "io/fs"

mfs := mockfs.NewStatFS(ctrl)
mfs.EXPECT().Stat("missing.txt").Return(nil, fs.ErrNotExist)
```

### Combining a mock FS with a mock File

```go
import "os"

data := []byte("content")
fsys := mockfs.NewOpenFileFS(ctrl)
file := mockfs.NewWriter(ctrl)

file.EXPECT().Write(gomock.Any()).Return(len(data), nil)
file.EXPECT().Close().Return(nil)
fsys.EXPECT().OpenFile("out.txt", os.O_WRONLY|os.O_CREATE, 0o644).Return(file, nil)
```

### Expecting multiple calls

```go
fs := mockfs.NewRemoveFS(ctrl)

// Expect exactly two calls
fs.EXPECT().Remove(gomock.Any()).Return(nil).Times(2)

// Expect at least one call
fs.EXPECT().Open(".").Return(mockfs.NewFile(ctrl), nil).MinTimes(1)

// Expect any number of calls (including zero)
fs.EXPECT().Stat(gomock.Any()).Return(nil, nil).AnyTimes()
```

### Dynamic responses with DoAndReturn

```go
import "os"

fs := mockfs.NewReadFileFS(ctrl)
fs.EXPECT().
    ReadFile(gomock.Any()).
    DoAndReturn(func(name string) ([]byte, error) {
        return os.ReadFile("testdata/" + name)
    })
```

## Available mocks

### Filesystem interfaces (`mock_fs.go`)

| Mock | Interface | Key method(s) |
|------|-----------|---------------|
| `NewChmodFS` | `ChmodFS` | `Chmod(name, mode)` |
| `NewChownFS` | `ChownFS` | `Chown(name, uid, gid)` |
| `NewChtimesFS` | `ChtimesFS` | `Chtimes(name, atime, mtime)` |
| `NewCopyFS` | `CopyFS` | `Copy(dir, fsys)` |
| `NewCreateFS` | `CreateFS` | `Create(name)` |
| `NewCreateTempFS` | `CreateTempFS` | `CreateTemp(dir, pattern)` |
| `NewGlobFS` | `GlobFS` | `Glob(pattern)` |
| `NewLinkerFS` | `LinkerFS` | `Symlink(oldname, newname)`, `ReadLink(name)` |
| `NewMkdirFS` | `MkdirFS` | `Mkdir(name, mode)` |
| `NewMkdirAllFS` | `MkdirAllFS` | `MkdirAll(name, mode)` |
| `NewMkdirTempFS` | `MkdirTempFS` | `MkdirTemp(dir, pattern)` |
| `NewOpenFileFS` | `OpenFileFS` | `OpenFile(name, flag, perm)` |
| `NewOsFS` | `OsFS` | Full OS filesystem |
| `NewReadDirFS` | `ReadDirFS` | `ReadDir(name)` |
| `NewReadDirNamesFS` | `ReadDirNamesFS` | `ReadDirNames(name)` |
| `NewReadFileFS` | `ReadFileFS` | `ReadFile(name)` |
| `NewReadLinkFS` | `ReadLinkFS` | `ReadLink(name)` |
| `NewRemoveFS` | `RemoveFS` | `Remove(name)` |
| `NewRemoveAllFS` | `RemoveAllFS` | `RemoveAll(name)` |
| `NewRenameFS` | `RenameFS` | `Rename(oldpath, newpath)` |
| `NewStatFS` | `StatFS` | `Stat(name)` |
| `NewSubFS` | `SubFS` | `Sub(dir)` |
| `NewSymlinkFS` | `SymlinkFS` | `Symlink(oldname, newname)` |
| `NewTempFileFS` | `TempFileFS` | `TempFile(dir, pattern)` |
| `NewWriteFileFS` | `WriteFileFS` | `WriteFile(name, data, perm)` |

All FS mocks also implement `Open(name)` from the base `fs.FS` interface.

### File and IO interfaces (`mock_file.go`)

| Mock | Interface | Key method(s) |
|------|-----------|---------------|
| `NewFile` | `File` | `Read`, `Stat`, `Close` |
| `NewDirNameReader` | `DirNameReader` | `ReadDirNames(n)` |
| `NewDirReader` | `DirReader` | `ReadDir(n)` |
| `NewOperation` | `Operation` | `Subject()` |
| `NewReaderAt` | `ReaderAt` | `ReadAt(p, off)` |
| `NewReadDirFile` | `ReadDirFile` | `ReadDir(n)` |
| `NewSeeker` | `Seeker` | `Seek(offset, whence)` |
| `NewStringWriter` | `StringWriter` | `WriteString(s)` |
| `NewSyncer` | `Syncer` | `Sync()` |
| `NewTruncater` | `Truncater` | `Truncate(size)` |
| `NewWriter` | `Writer` | `Write(p)` |
| `NewWriterAt` | `WriterAt` | `WriteAt(p, off)` |

All file mocks implement the base `File` interface (`Read`, `Stat`, `Close`).

## Regenerating

Mocks are generated from `generate.sh` using [mockgen](https://github.com/uber-go/mock):

```
make generate
```
