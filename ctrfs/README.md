# ctrfs

[![Go Reference](https://pkg.go.dev/badge/github.com/unstoppablemango/ihfs/ctrfs.svg)](https://pkg.go.dev/github.com/unstoppablemango/ihfs/ctrfs)
[![Go version](https://img.shields.io/github/go-mod/go-version/UnstoppableMango/ihfs?filename=ctrfs/go.mod)](https://github.com/UnstoppableMango/ihfs/blob/main/ctrfs/go.mod)
[![Codecov](https://img.shields.io/codecov/c/github/UnstoppableMango/ihfs?flag=ctrfs&label=coverage)](https://codecov.io/gh/UnstoppableMango/ihfs?flags[0]=ctrfs)

An `io/fs` adapter for OCI container images and layers, backed by [google/go-containerregistry](https://github.com/google/go-containerregistry).

## Usage

### Reading a full image

`FromImage` merges all layers (with whiteouts applied) into a single read-only filesystem:

```go
import (
    "github.com/google/go-containerregistry/pkg/v1/remote"
    "github.com/unstoppablemango/ihfs/ctrfs"
)

img, err := remote.Image(ref)
fsys := ctrfs.FromImage(img)
defer fsys.Close()

f, err := fsys.Open("etc/os-release")
```

### Reading a single layer

`FromLayer` exposes one layer's uncompressed tar stream as a filesystem:

```go
layer, err := img.LayerByDiffID(hash)
fsys, err := ctrfs.FromLayer(layer)
defer fsys.Close()

data, err := fs.ReadFile(fsys, "usr/bin/myapp")
```

### Writing to a layer or image

`ToLayer` walks an `fs.FS` rooted at `dir` and produces an OCI layer:

```go
layer, err := ctrfs.ToLayer(myFS, ".")
```

`ToImage` appends that layer onto a base image in one step:

```go
newImg, err := ctrfs.ToImage(baseImg, myFS, ".")
```

To root the layer at a subdirectory:

```go
layer, err := ctrfs.ToLayer(myFS, "dist")
```
