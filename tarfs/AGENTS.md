# tarfs

This package contains two complementary implementations:

- **`TarFile`** — read-only filesystem backed by a tar archive with lazy, streaming reads
- **`Writer`** — write-only tar-backed filesystem supporting `Create`, `Mkdir`, `OpenFile`, and `Copy`

## Design

### TarFile (read-only)

- Tar entries are read on-demand and cached in a map protected by a `sync.RWMutex`
- Synthetic directory entries are created for paths present in the archive but not explicitly listed as directories
- Opening a directory forces a full drain of the tar stream to build a complete directory listing
- `File.ReadDir` builds a stable snapshot on first call to support consistent pagination

### Writer (write-only)

- Wraps an underlying `tar.Writer` to write entries into a tar archive
- Supports `Create`, `Mkdir`, `OpenFile`, and `Copy` operations

## Key Constraints

- **`TarFile` is read-only**: no write operations are supported on the reader
- **`Writer` is write-only**: no read operations are supported on the writer
- Opening later entries may require draining the tar stream until that entry is reached; cached entries then allow random access to previously seen entries
- Memory for entries is cached and may remain referenced by open `File` handles even after `TarFile.Close()` (which only closes the underlying stream); memory is released only once no references remain and GC runs, so avoid holding files open unnecessarily on large archives

## Relationships

- Uses `osfs.Default` internally when opening archives by path

## Coverage

Aim for high coverage, but do not write messy or low-value tests just to hit 100%. If a branch requires complex setup with little benefit, skip it.
