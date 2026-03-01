// https://github.com/spf13/afero/blob/master/const_bsds.go

//go:build aix || darwin || freebsd || openbsd || netbsd || dragonfly || zos || solaris

package union

import "syscall"

// BADFD is the "bad file descriptor" error code for BSD platforms.
const BADFD = syscall.EBADF
