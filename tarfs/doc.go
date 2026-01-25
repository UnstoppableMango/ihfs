// Package tarfs provides a read-only file system interface to tar archives.
// It will lazily buffer the contents of the tar archive as files are accessed.
//
// Entries are accessed in order and cached as they are read, so random access may be inefficient.
package tarfs
