// Package corfs implements a cache-on-read filesystem. When files are read from
// the base filesystem, they are cached in the layer. Future reads come from the
// cached version until it expires or is invalidated.
//
// The implementation is based heavily on [afero.CacheOnReadFs].
package corfs
