# tarfs

Read-only filesystem backed by a tar archive with lazy, streaming reads.

## Design

- Tar entries are read on-demand and cached in a map protected by a `sync.RWMutex`
- Synthetic directory entries are created for paths present in the archive but not explicitly listed as directories
- Opening a directory forces a full drain of the tar stream to build a complete directory listing
- `File.ReadDir` builds a stable snapshot on first call to support consistent pagination

## Key Constraints

- **Read-only**: no write operations are supported
- Random access requires seeking back to the beginning of the archive (expensive if the underlying reader is not seekable)
- Memory is held in the entry cache until the `TarFile` is closed — avoid holding open unnecessarily on large archives

## Relationships

- Uses `osfs.Default` internally when opening archives by path

## Coverage

100% coverage required.
