# corfs

Cache-on-read filesystem: files are copied from base to layer on first read; subsequent reads come from the cache (layer).

## Design

- Based on afero's `CacheOnReadFs`
- Cache state is one of: `cacheMiss`, `cacheStale`, `cacheHit`, `cacheLocal`
- Staleness is determined by comparing modtimes; TTL of 0 means cache never expires
- Directories are always merged (never fully cached) — only regular files are cached

## Key Behaviors

- On a cache miss or stale hit, the file is copied to layer via `union.CopyToLayer`
- Writes go to **both** base and layer simultaneously
- Cache TTL is configurable via `corfs.WithCacheTime(d time.Duration)` option

## Relationships

- Uses `union.File` for merged directory reads
- Uses `union.CopyToLayer` when caching a file

## Coverage

100% coverage required.
