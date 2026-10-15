package runtime

import (
	"syscall"
	"unsafe"
)

const (
	sockCloexec = 0
	msgNoSignal = 0
)

func unsafePointer(p *byte) unsafe.Pointer { return unsafe.Pointer(p) }

func accept4(fd int) (int, syscall.Sockaddr, error) {
	nfd, sa, err := syscall.Accept(fd)
	if err == nil {
		syscall.CloseOnExec(nfd)
	}
	return nfd, sa, err
}

func sendNoPipe(fd int, data []byte, flags int) (int, error) {
	return syscall.Write(fd, data)
}

func setsockoptBytes(fd, level, opt int, v []byte) error { return syscall.ENOPROTOOPT }

func getsockoptBytes(fd, level, opt, size int) ([]byte, error) { return nil, syscall.ENOPROTOOPT }
