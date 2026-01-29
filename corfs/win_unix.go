// https://github.com/spf13/afero/blob/master/const_win_unix.go

//go:build !aix && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly && !zos && !solaris

package corfs

import "syscall"

const BADFD = syscall.EBADFD
