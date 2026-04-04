# ghfs

Read-only filesystem backed by the GitHub API (releases and release assets).

## Path Format

Paths follow one of three conventions.

- api.github.com API paths
- github.com Web paths
- raw.githubusercontent.com Web paths

See `path.go` for the full grammar parsed by `Parse`.

## Design

- `Fs.Open(".")` returns a synthetic root directory
- Non-`.` paths are parsed by `Parse`; an invalid path returns `ErrInvalid`
- Asset lookup: if an asset name is given, the release is fetched first to resolve the asset ID, then the asset download URL is used
- HTTP requests go through `github.Client.BareDo` to get a raw `io.ReadCloser`
- Context injection: a `ContextFunc` (default: `context.Background`) is called per-operation, allowing callers to attach deadlines or cancellation

## Key Constraints

- **Read-only**: only `Open` is implemented; all other FS interfaces return `ErrNotImplemented`
- The package has its own `go.mod` (separate module) under `ghfs/`; update it independently

## Testing

- Use `ghfs_test` (external) for public API tests
- Internal tests (`fs_internal_test.go`) test unexported helpers
- Mock the `github.Client` rather than hitting the real API

## Coverage

Aim for high coverage, but do not write messy or low-value tests just to hit 100%. If a branch requires complex setup with little benefit, skip it.
