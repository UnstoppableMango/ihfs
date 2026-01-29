// https://github.com/spf13/afero/blob/master/const_win_unix.go

//go:build !darwin && !openbsd && !freebsd && !dragonfly && !netbsd

package corfs

import "syscall"

const BADFD = syscall.EBADFD
