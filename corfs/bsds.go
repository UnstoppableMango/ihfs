// https://github.com/spf13/afero/blob/master/const_bsds.go

//go:build aix || darwin || freebsd || openbsd || netbsd || dragonfly || zos || solaris

package corfs

import "syscall"

const BADFD = syscall.EBADF
