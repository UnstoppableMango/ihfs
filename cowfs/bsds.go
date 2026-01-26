//go:build aix || darwin || freebsd || openbsd || netbsd || dragonfly || zos || solaris

package cowfs

import "syscall"

// https://github.com/spf13/afero/blob/master/const_bsds.go

const BADFD = syscall.EBADF
