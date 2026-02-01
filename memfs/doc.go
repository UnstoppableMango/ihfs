// Package memfs provides an in-memory filesystem implementation.
//
// The in-memory filesystem (MemFS) stores all file data in memory,
// making it ideal for testing, temporary storage, or scenarios where
// persistence is not required.
//
// # Features
//
//   - Thread-safe operations using mutexes
//   - Full filesystem operations (create, read, write, delete, etc.)
//   - Directory hierarchy support
//   - File metadata (permissions, timestamps, ownership)
//   - No third-party dependencies beyond ihfs and the standard library
//
// # Example Usage
//
//	fs := memfs.New()
//
//	// Create a file
//	file, err := fs.Create("/hello.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer file.Close()
//
//	// Write data
//	_, err = file.Write([]byte("Hello, World!"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Implementation Notes
//
// This implementation is based on afero's MemMapFs but adapted to work
// with ihfs interfaces. All file data is stored in a map with path keys,
// and operations are synchronized using a read-write mutex.
package memfs
