// https://github.com/spf13/afero/blob/master/const_win_unix.go

//go:build !aix && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly && !zos && !solaris

package union

import "syscall"

// BADFD is the "bad file descriptor" error code for non-Unix platforms.
const BADFD = syscall.EBADFD
