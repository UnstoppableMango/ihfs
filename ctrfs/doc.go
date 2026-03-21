// Module ctrfs provides fs.FS implementations and write helpers for OCI container images and layers.
//
// Sub-packages:
//   - [github.com/unstoppablemango/ihfs/ctrfs/v1/image] – read-only FS from a [v1.Image] and helper to append a new layer
//   - [github.com/unstoppablemango/ihfs/ctrfs/v1/layer] – read-only FS from a [v1.Layer] and helper to create a new layer
package ctrfs
