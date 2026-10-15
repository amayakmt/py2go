package runtime

import "syscall"

func sysSelect(n int, r, w, x *syscall.FdSet, tv *syscall.Timeval) error {
	_, err := syscall.Select(n, r, w, x, tv)
	return err
}
