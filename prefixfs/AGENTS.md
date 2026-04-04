# prefixfs

Makes an underlying filesystem accessible only under a prefix path. The inverse of `fs.Sub`.

## Design

- `New(fsys, prefix)` panics if the cleaned prefix is not a valid, non-root `fs.ValidPath`
- Opening `"."` or any ancestor of the prefix returns a synthetic `dir` that lists only the next path component
- Opening exactly the prefix returns the underlying FS root (`"."`)
- Opening a path under the prefix strips the prefix and delegates to the underlying FS

## Key Types

- `dir`: synthetic read-only directory returned for paths above the prefix boundary
- `dirEntry` / `dirInfo`: minimal `fs.DirEntry` / `fs.FileInfo` for synthetic directories

## Relationships

- Pure `io/fs` package — does not depend on `ihfs` types; usable with any `fs.FS`

## Coverage

Aim for high coverage, but do not write messy or low-value tests just to hit 100%. If a branch requires complex setup with little benefit, skip it.
