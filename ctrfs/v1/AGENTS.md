# ctrfs/v1

`io/fs` adapters for OCI v1 images and layers.

## Design

- `Layer` wraps an OCI v1 layer and exposes its contents as a read-only `fs.FS`
- `Image` wraps an OCI v1 image and presents its merged layer stack as a single `fs.FS`
- Whiteout files (`.wh.` prefix) are handled during directory listing — deleted entries are suppressed

## Key Constraints

- **Read-only**: OCI layers are immutable; no write operations are supported
- This package lives in its own Go module (`ctrfs/go.mod`); dependency updates must be managed separately from the root module
- Run `gomod2nix generate` inside `ctrfs/` after updating `go.mod`
