package runtime

import "syscall"

func sysSelect(n int, r, w, x *syscall.FdSet, tv *syscall.Timeval) error {
	return syscall.Select(n, r, w, x, tv)
}
