# cowfs

Copy-on-write filesystem: reads come from both layers (layer preferred), writes go to the overlay layer only.

## Design

- Based on afero's `CopyOnWriteFs`
- Layer is checked first; base is the fallback for reads
- On first write to a base-only file, the file is copied to layer before modification
- Directory listings merge entries from both layers via `union.MergeStrategy`

## Key Behaviors

- `Open`: returns layer file if present, else base file; for directories, returns a merged `union.File`
- `Stat`: prefers layer, falls back to base
- Write operations (`Create`, `OpenFile`, etc.) always target the layer
- `Remove`/`Rename`: only affect the layer — base files remain untouched

## Relationships

- Delegates directory merging to `union.File`
- Pass `union.Option` values through `cowfs.New` to configure merge strategy
