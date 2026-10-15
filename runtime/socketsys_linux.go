package runtime

import (
	"syscall"
	"unsafe"
)

func unsafePointer(p *byte) unsafe.Pointer { return unsafe.Pointer(p) }

func sendNoPipe(fd int, data []byte, flags int) (int, error) {
	n, _, e := syscall.Syscall6(syscall.SYS_SENDTO, uintptr(fd), uintptr(sliceBase(data)), uintptr(len(data)), uintptr(flags), 0, 0)
	if e != 0 {
		return 0, e
	}
	return int(n), nil
}

func sliceBase(b []byte) unsafe.Pointer {
	if len(b) == 0 {
		return nil
	}
	return unsafe.Pointer(&b[0])
}

func setsockoptBytes(fd, level, opt int, v []byte) error {
	var p unsafe.Pointer
	if len(v) > 0 {
		p = unsafe.Pointer(&v[0])
	}
	_, _, e := syscall.Syscall6(syscall.SYS_SETSOCKOPT, uintptr(fd), uintptr(level), uintptr(opt), uintptr(p), uintptr(len(v)), 0)
	if e != 0 {
		return e
	}
	return nil
}

func getsockoptBytes(fd, level, opt, size int) ([]byte, error) {
	b := make([]byte, size)
	n := uint32(size)
	var p unsafe.Pointer
	if size > 0 {
		p = unsafe.Pointer(&b[0])
	}
	_, _, e := syscall.Syscall6(syscall.SYS_GETSOCKOPT, uintptr(fd), uintptr(level), uintptr(opt), uintptr(p), uintptr(unsafe.Pointer(&n)), 0)
	if e != 0 {
		return nil, e
	}
	return b[:n], nil
}

const (
	sockCloexec = syscall.SOCK_CLOEXEC
	msgNoSignal = syscall.MSG_NOSIGNAL
)

func accept4(fd int) (int, syscall.Sockaddr, error) { return syscall.Accept4(fd, syscall.SOCK_CLOEXEC) }
