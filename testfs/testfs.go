package testfs

import "errors"

// ErrNotImplemented is returned when a mock method has not been configured.
var ErrNotImplemented = errors.New("mock method not implemented")
