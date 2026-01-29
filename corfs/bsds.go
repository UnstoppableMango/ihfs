// https://github.com/spf13/afero/blob/master/const_bsds.go

//go:build darwin || openbsd || freebsd || dragonfly || netbsd

package corfs

import "syscall"

const BADFD = syscall.EBADF
