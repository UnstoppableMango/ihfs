# memfs - In-Memory Filesystem

An in-memory filesystem implementation for the ihfs library, based on afero's MemMapFs.

## Features

- **Thread-safe**: All operations are synchronized using mutexes
- **Full filesystem support**: Create, read, write, delete, rename, etc.
- **Directory hierarchy**: Complete directory tree support
- **File metadata**: Permissions, timestamps, ownership
- **No persistence**: All data stored in memory (ideal for testing)

## Installation

```bash
go get github.com/unstoppablemango/ihfs/memfs
```

## Usage

```go
package main

import (
    "fmt"
    "io"
    "log"
    
    "github.com/unstoppablemango/ihfs/memfs"
)

func main() {
    // Create a new in-memory filesystem
    fs := memfs.New()
    
    // Create a file
    file, err := fs.Create("/hello.txt")
    if err != nil {
        log.Fatal(err)
    }
    
    // Write data
    writer := file.(io.Writer)
    _, err = writer.Write([]byte("Hello, World!"))
    if err != nil {
        log.Fatal(err)
    }
    file.Close()
    
    // Read the file back
    file, err = fs.Open("/hello.txt")
    if err != nil {
        log.Fatal(err)
    }
    
    content, err := io.ReadAll(file)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(content)) // Output: Hello, World!
}
```

## Features

### File Operations

- `Create(name string)` - Create a new file
- `Open(name string)` - Open a file for reading
- `OpenFile(name string, flag int, perm os.FileMode)` - Open with flags
- `Remove(name string)` - Remove a file or empty directory
- `RemoveAll(name string)` - Remove a file or directory and all contents
- `Rename(oldName, newName string)` - Rename a file or directory
- `Stat(name string)` - Get file information

### Directory Operations

- `Mkdir(name string, perm os.FileMode)` - Create a directory
- `MkdirAll(name string, perm os.FileMode)` - Create nested directories
- `ReadDir(n int)` - Read directory contents (via file handle)

### Metadata Operations

- `Chmod(name string, mode os.FileMode)` - Change file permissions
- `Chown(name string, uid, gid int)` - Change file ownership
- `Chtimes(name string, atime, mtime time.Time)` - Change file times

### File Handle Operations

- `Read(p []byte)` - Read from file
- `Write(p []byte)` - Write to file
- `Seek(offset int64, whence int)` - Seek to position
- `Truncate(size int64)` - Truncate file to size
- `Close()` - Close file handle

## Implementation Details

### Based on afero.MemMapFs

This implementation is heavily inspired by [afero's MemMapFs](https://github.com/spf13/afero), but adapted to work with ihfs interfaces:

- Uses `sync.RWMutex` for thread-safe operations
- Stores all data in a `map[string]*FileData`
- Root directory (`/`) is created automatically
- Paths are normalized to use consistent separators

### Memory Structure

```
Fs
├── mu (RWMutex)
└── data (map[string]*FileData)
    ├── "/" (root directory)
    ├── "/file.txt" (file)
    └── "/dir" (directory)
        ├── children (map)
        └── "/dir/nested.txt"
```

### Thread Safety

All filesystem operations acquire appropriate locks:
- Read operations use `RLock()` 
- Write operations use `Lock()`
- File data has its own mutex for content operations

## Testing

The package includes comprehensive tests using Ginkgo:

```bash
go test ./memfs/...
```

## Comparison with afero.MemMapFs

| Feature | memfs | afero.MemMapFs |
|---------|-------|----------------|
| Interface | ihfs.FS | afero.Fs |
| Thread Safety | ✅ | ✅ |
| Directory Support | ✅ | ✅ |
| Permissions | ✅ | ✅ |
| Ownership (uid/gid) | ✅ | ✅ |
| Timestamps | ✅ | ✅ |
| Symlinks | ❌ | ✅ |
| Hard Links | ❌ | ❌ |

## License

Copyright © 2024

Licensed under the Apache License, Version 2.0.

## Credits

Based on [afero's MemMapFs](https://github.com/spf13/afero) by Steve Francia and contributors.
