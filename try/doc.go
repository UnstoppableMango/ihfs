// Package try provides strict filesystem operation wrappers that require
// the underlying FS to implement the specific interface for each operation.
//
// Unlike the convenience functions in the root ihfs package (e.g., [ihfs.Stat],
// [ihfs.MkdirAll]), which may use fallback strategies when the FS lacks a
// specific interface, functions in this package return [ErrNotImplemented] if
// the required interface is not satisfied. This makes them useful when you need
// to know whether a capability exists rather than silently falling back.
//
// For example, [ihfs.Stat] delegates to [fs.Stat] which may open the file and
// call Stat on the resulting handle, while [Stat] returns [ErrNotImplemented]
// if the FS does not implement [ihfs.StatFS].
package try
