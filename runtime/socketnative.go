package runtime

import (
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// The natives of the socket module: BSD sockets on raw descriptors, with the
// timeout the program set handled the way CPython handles it (a socket with a
// timeout is non-blocking and waits for readiness), and name resolution by
// Go's resolver.

func sockAddrIPv4(host string) ([4]byte, error) {
	var out [4]byte
	switch host {
	case "":
		return out, nil
	case "<broadcast>":
		return [4]byte{255, 255, 255, 255}, nil
	case "<any>":
		return out, nil
	}
	if ip := net.ParseIP(host); ip != nil && ip.To4() != nil {
		copy(out[:], ip.To4())
		return out, nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return out, gaiError(err, host)
	}
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			copy(out[:], v4)
			return out, nil
		}
	}
	return out, gaiErrorCode(-9, "Address family for hostname not supported")
}

func sockAddrIPv6(host string) ([16]byte, uint32, error) {
	var out [16]byte
	zone := uint32(0)
	if i := strings.IndexByte(host, '%'); i >= 0 {
		if ifc, err := net.InterfaceByName(host[i+1:]); err == nil {
			zone = uint32(ifc.Index)
		} else if n, err := strconv.Atoi(host[i+1:]); err == nil {
			zone = uint32(n)
		}
		host = host[:i]
	}
	if host == "" || host == "<any>" {
		return out, zone, nil
	}
	if host == "<broadcast>" {
		return out, zone, gaiErrorCode(-9, "Address family for hostname not supported")
	}
	if ip := net.ParseIP(host); ip != nil {
		copy(out[:], ip.To16())
		return out, zone, nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return out, zone, gaiError(err, host)
	}
	for _, ip := range ips {
		if ip.To4() == nil {
			copy(out[:], ip.To16())
			return out, zone, nil
		}
	}
	return out, zone, gaiErrorCode(-9, "Address family for hostname not supported")
}

func gaiErrorCode(code int, msg string) *PyError {
	e := &PyError{Type: "socket_pm_gaierror", Args: []any{int64(code), msg}}
	e.Attrs = map[string]any{"errno": int64(code), "strerror": msg, "filename": nil, "filename2": nil}
	return e
}

func gaiError(err error, host string) *PyError {
	if de, ok := err.(*net.DNSError); ok {
		if de.IsTemporary || de.IsTimeout {
			return gaiErrorCode(-3, "Temporary failure in name resolution")
		}
	}
	return gaiErrorCode(-2, "Name or service not known")
}

func portArg(v any) int {
	n, ok := intOf(Unbox(v))
	if !ok {
		panic(newError("TypeError", "'"+TypeName(Unbox(v))+"' object cannot be interpreted as an integer"))
	}
	if n < 0 || n > 65535 {
		panic(overflowError("bind(): port must be 0-65535."))
	}
	return int(n)
}

// toSockaddr turns a Python address into a sockaddr for the family.
func toSockaddr(family int, addr any, fn string) syscall.Sockaddr {
	addr = Unbox(addr)
	switch family {
	case syscall.AF_INET:
		t, ok := addr.(Seq)
		if !ok || t.Len() != 2 {
			if ok {
				panic(newError("TypeError", "AF_INET address must be tuple, not "+TypeName(addr)))
			}
			panic(newError("TypeError", "AF_INET address must be tuple, not "+TypeName(addr)))
		}
		host := hostArg(t.At(0))
		ip, err := sockAddrIPv4(host)
		if err != nil {
			panic(err)
		}
		return &syscall.SockaddrInet4{Port: portArg(t.At(1)), Addr: ip}
	case syscall.AF_INET6:
		t, ok := addr.(Seq)
		if !ok || t.Len() < 2 || t.Len() > 4 {
			panic(newError("TypeError", "AF_INET6 address must be tuple, not "+TypeName(addr)))
		}
		host := hostArg(t.At(0))
		ip, zone, err := sockAddrIPv6(host)
		if err != nil {
			panic(err)
		}
		if t.Len() == 4 {
			zone = uint32(DynInt(t.At(3)))
		}
		return &syscall.SockaddrInet6{Port: portArg(t.At(1)), Addr: ip, ZoneId: zone}
	case syscall.AF_UNIX:
		switch p := addr.(type) {
		case string:
			return unixAddr(p)
		case Bytes:
			return unixAddr(string(p))
		}
		panic(newError("TypeError", "a bytes-like object is required, not '"+TypeName(addr)+"'"))
	}
	panic(sysError(syscall.EAFNOSUPPORT))
}

func unixAddr(p string) syscall.Sockaddr {
	if strings.IndexByte(p, 0) == 0 {
		return &syscall.SockaddrUnix{Name: "@" + p[1:]}
	}
	return &syscall.SockaddrUnix{Name: p}
}

func hostArg(v any) string {
	switch h := Unbox(v).(type) {
	case string:
		return h
	case Bytes:
		return string(h)
	}
	panic(newError("TypeError", "str, bytes or bytearray expected, not "+TypeName(Unbox(v))))
}

func fromSockaddr(sa syscall.Sockaddr) any {
	switch a := sa.(type) {
	case *syscall.SockaddrInet4:
		return NewVarTuple[any](net.IP(a.Addr[:]).String(), int64(a.Port))
	case *syscall.SockaddrInet6:
		host := net.IP(a.Addr[:]).String()
		if a.ZoneId != 0 && strings.HasPrefix(host, "fe80") {
			if ifc, err := net.InterfaceByIndex(int(a.ZoneId)); err == nil {
				host += "%" + ifc.Name
			}
		}
		return NewVarTuple[any](host, int64(a.Port), int64(0), int64(a.ZoneId))
	case *syscall.SockaddrUnix:
		if a.Name == "@" {
			return ""
		}
		if strings.HasPrefix(a.Name, "@") {
			return Bytes("\x00" + a.Name[1:])
		}
		return a.Name
	}
	return nil
}

// waitFd waits until the descriptor is ready or the deadline passes; false
// when it passed. timeout < 0 waits for ever.
func waitFd(fd int, write bool, timeout float64) bool {
	if fd >= 1024 {
		return true
	}
	var set syscall.FdSet
	set.Bits[fd/64] |= 1 << (uint(fd) % 64)
	var tv *syscall.Timeval
	if timeout >= 0 {
		t := syscall.NsecToTimeval(int64(timeout * 1e9))
		tv = &t
	}
	var ready bool
	Blocking(func() {
		for {
			var err error
			if write {
				err = sysSelect(fd+1, nil, &set, nil, tv)
			} else {
				err = sysSelect(fd+1, &set, nil, nil, tv)
			}
			if err == syscall.EINTR {
				continue
			}
			ready = err == nil && set.Bits[fd/64]&(1<<(uint(fd)%64)) != 0
			return
		}
	})
	return ready
}

func timeoutArg(args []any, i int) float64 {
	if i >= len(args) || IsNone(args[i]) {
		return -1
	}
	return mathFloat(args[i])
}

func timedOut() *PyError { return newError("TimeoutError", "timed out") }

// retryIO runs op until it does not fail with EAGAIN: with a timeout it
// waits for readiness meanwhile, and raises TimeoutError when time is up.
func retryIO(fd int, write bool, timeout float64, op func() error) error {
	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(time.Duration(timeout * float64(time.Second)))
	}
	for {
		var err error
		if timeout < 0 {
			Blocking(func() {
				for {
					err = op()
					if err != syscall.EINTR {
						break
					}
				}
			})
		} else {
			err = op()
		}
		if err == syscall.EINTR {
			continue
		}
		if err != syscall.EAGAIN || timeout == 0 {
			return err
		}
		remaining := -1.0
		if !deadline.IsZero() {
			remaining = time.Until(deadline).Seconds()
			if remaining <= 0 {
				panic(timedOut())
			}
		}
		if !waitFd(fd, write, remaining) {
			panic(timedOut())
		}
	}
}

func setNonblock(fd int, on bool) {
	syscall.SetNonblock(fd, on)
}

func init() {
	RegisterNative("_socket.socket", func(args []any, kw []KW) any {
		fam, typ, proto := int(DynInt(args[0])), int(DynInt(args[1])), int(DynInt(args[2]))
		fd, err := syscall.Socket(fam, typ|sockCloexec, proto)
		if err != nil {
			panic(sysError(err))
		}
		syscall.CloseOnExec(fd)
		return int64(fd)
	})
	RegisterNative("_socket.socketpair", func(args []any, kw []KW) any {
		fam, typ, proto := int(DynInt(args[0])), int(DynInt(args[1])), int(DynInt(args[2]))
		fds, err := syscall.Socketpair(fam, typ|sockCloexec, proto)
		if err != nil {
			panic(sysError(err))
		}
		return NewVarTuple[any](int64(fds[0]), int64(fds[1]))
	})
	RegisterNative("_socket.close", func(args []any, kw []KW) any {
		if err := syscall.Close(int(DynInt(args[0]))); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("_socket.set_nonblock", func(args []any, kw []KW) any {
		setNonblock(int(DynInt(args[0])), Truth(args[1]))
		return nil
	})
	RegisterNative("_socket.shutdown", func(args []any, kw []KW) any {
		if err := syscall.Shutdown(int(DynInt(args[0])), int(DynInt(args[1]))); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("_socket.bind", func(args []any, kw []KW) any {
		fd, fam := int(DynInt(args[0])), int(DynInt(args[1]))
		sa := toSockaddr(fam, args[2], "bind")
		if err := syscall.Bind(fd, sa); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("_socket.listen", func(args []any, kw []KW) any {
		backlog := 128
		if len(args) > 1 && !IsNone(args[1]) {
			backlog = int(DynInt(args[1]))
		}
		if backlog < 0 {
			backlog = 0
		}
		if err := syscall.Listen(int(DynInt(args[0])), backlog); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	// _socket.connect(fd, family, addr, timeout)
	RegisterNative("_socket.connect", func(args []any, kw []KW) any {
		fd, fam := int(DynInt(args[0])), int(DynInt(args[1]))
		sa := toSockaddr(fam, args[2], "connect")
		timeout := timeoutArg(args, 3)
		err := syscall.Connect(fd, sa)
		if err == syscall.EINTR {
			err = syscall.EINPROGRESS
		}
		if err == syscall.EINPROGRESS && timeout != 0 {
			ok := waitFd(fd, true, timeout)
			if !ok {
				panic(timedOut())
			}
			soerr, gerr := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_ERROR)
			if gerr != nil {
				panic(sysError(gerr))
			}
			if soerr != 0 {
				panic(sysError(syscall.Errno(soerr)))
			}
			return nil
		}
		if err != nil {
			panic(sysError(err))
		}
		return nil
	})
	// _socket.accept(fd, timeout) -> (fd, address)
	RegisterNative("_socket.accept", func(args []any, kw []KW) any {
		fd := int(DynInt(args[0]))
		timeout := timeoutArg(args, 1)
		var nfd int
		var sa syscall.Sockaddr
		err := retryIO(fd, false, timeout, func() (err error) {
			nfd, sa, err = accept4(fd)
			return
		})
		if err != nil {
			panic(sysError(err))
		}
		return NewVarTuple[any](int64(nfd), fromSockaddr(sa))
	})
	// _socket.recv(fd, n, flags, timeout) -> bytes
	RegisterNative("_socket.recv", func(args []any, kw []KW) any {
		fd, n := int(DynInt(args[0])), int(DynInt(args[1]))
		if n < 0 {
			panic(valueError("negative buffersize in recv"))
		}
		flags := int(intArgAt(args, 2, 0))
		timeout := timeoutArg(args, 3)
		buf := make([]byte, n)
		var got int
		err := retryIO(fd, false, timeout, func() (err error) {
			got, _, err = syscall.Recvfrom(fd, buf, flags)
			return
		})
		if err != nil {
			panic(sysError(err))
		}
		return Bytes(buf[:got])
	})
	RegisterNative("_socket.recvfrom", func(args []any, kw []KW) any {
		fd, n := int(DynInt(args[0])), int(DynInt(args[1]))
		flags := int(intArgAt(args, 2, 0))
		timeout := timeoutArg(args, 3)
		buf := make([]byte, n)
		var got int
		var from syscall.Sockaddr
		err := retryIO(fd, false, timeout, func() (err error) {
			got, from, err = syscall.Recvfrom(fd, buf, flags)
			return
		})
		if err != nil {
			panic(sysError(err))
		}
		return NewVarTuple[any](Bytes(buf[:got]), fromSockaddr(from))
	})
	// _socket.send(fd, data, flags, timeout) -> sent
	RegisterNative("_socket.send", func(args []any, kw []KW) any {
		fd := int(DynInt(args[0]))
		data := bytesArg(args[1])
		flags := int(intArgAt(args, 2, 0)) | msgNoSignal
		timeout := timeoutArg(args, 3)
		var sent int
		err := retryIO(fd, true, timeout, func() (err error) {
			sent, err = sendNoPipe(fd, data, flags)
			return
		})
		if err != nil {
			panic(sysError(err))
		}
		return int64(sent)
	})
	// _socket.sendto(fd, data, flags, family, addr, timeout)
	RegisterNative("_socket.sendto", func(args []any, kw []KW) any {
		fd := int(DynInt(args[0]))
		data := bytesArg(args[1])
		flags := int(intArgAt(args, 2, 0)) | msgNoSignal
		sa := toSockaddr(int(DynInt(args[3])), args[4], "sendto")
		timeout := timeoutArg(args, 5)
		err := retryIO(fd, true, timeout, func() error {
			return syscall.Sendto(fd, data, flags, sa)
		})
		if err != nil {
			panic(sysError(err))
		}
		return int64(len(data))
	})
	RegisterNative("_socket.getsockname", func(args []any, kw []KW) any {
		sa, err := syscall.Getsockname(int(DynInt(args[0])))
		if err != nil {
			panic(sysError(err))
		}
		return fromSockaddr(sa)
	})
	RegisterNative("_socket.getpeername", func(args []any, kw []KW) any {
		sa, err := syscall.Getpeername(int(DynInt(args[0])))
		if err != nil {
			panic(sysError(err))
		}
		return fromSockaddr(sa)
	})
	// _socket.setsockopt(fd, level, option, value): value is an int or bytes
	RegisterNative("_socket.setsockopt", func(args []any, kw []KW) any {
		fd, level, opt := int(DynInt(args[0])), int(DynInt(args[1])), int(DynInt(args[2]))
		var err error
		switch v := Unbox(args[3]).(type) {
		case Bytes:
			err = setsockoptBytes(fd, level, opt, []byte(v))
		case *ByteArray:
			err = setsockoptBytes(fd, level, opt, v.b)
		default:
			n := DynInt(v)
			switch {
			case level == syscall.SOL_SOCKET && (opt == syscall.SO_RCVTIMEO || opt == syscall.SO_SNDTIMEO):
				err = syscall.SetsockoptTimeval(fd, level, opt, &syscall.Timeval{Sec: n})
			default:
				err = syscall.SetsockoptInt(fd, level, opt, int(n))
			}
		}
		if err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("_socket.getsockopt", func(args []any, kw []KW) any {
		fd, level, opt := int(DynInt(args[0])), int(DynInt(args[1])), int(DynInt(args[2]))
		if len(args) > 3 && !IsNone(args[3]) {
			size := int(DynInt(args[3]))
			if size < 0 || size > 1024 {
				panic(newError("OSError", "getsockopt buflen out of range"))
			}
			out, err := getsockoptBytes(fd, level, opt, size)
			if err != nil {
				panic(sysError(err))
			}
			return Bytes(out)
		}
		v, err := syscall.GetsockoptInt(fd, level, opt)
		if err != nil {
			panic(sysError(err))
		}
		return int64(v)
	})
	// _socket.sendmsg(fd, buffers, ancdata, flags, family, addr, timeout)
	RegisterNative("_socket.sendmsg", func(args []any, kw []KW) any {
		fd := int(DynInt(args[0]))
		var data []byte
		for c := DynIter(args[1]).Iter(); c.Next(); {
			data = append(data, bytesArg(c.Value())...)
		}
		var oob []byte
		for c := DynIter(args[2]).Iter(); c.Next(); {
			item := c.Value().(Seq)
			level, typ := int(DynInt(item.At(0))), int(DynInt(item.At(1)))
			payload := bytesArg(item.At(2))
			oob = append(oob, controlMessage(level, typ, payload)...)
		}
		flags := int(intArgAt(args, 3, 0)) | msgNoSignal
		var to syscall.Sockaddr
		if len(args) > 5 && !IsNone(args[5]) {
			to = toSockaddr(int(DynInt(args[4])), args[5], "sendmsg")
		}
		timeout := timeoutArg(args, 6)
		var n int
		err := retryIO(fd, true, timeout, func() (err error) {
			n, err = syscall.SendmsgN(fd, data, oob, to, flags)
			return
		})
		if err != nil {
			panic(sysError(err))
		}
		return int64(n)
	})
	// _socket.recvmsg(fd, bufsize, ancbufsize, flags, timeout) -> (data, ancdata, flags, address)
	RegisterNative("_socket.recvmsg", func(args []any, kw []KW) any {
		fd := int(DynInt(args[0]))
		buf := make([]byte, int(DynInt(args[1])))
		oob := make([]byte, int(intArgAt(args, 2, 0)))
		flags := int(intArgAt(args, 3, 0))
		timeout := timeoutArg(args, 4)
		var n, oobn, rflags int
		var from syscall.Sockaddr
		err := retryIO(fd, false, timeout, func() (err error) {
			n, oobn, rflags, from, err = syscall.Recvmsg(fd, buf, oob, flags)
			return
		})
		if err != nil {
			panic(sysError(err))
		}
		anc := NewList[any]()
		if msgs, perr := syscall.ParseSocketControlMessage(oob[:oobn]); perr == nil {
			for _, m := range msgs {
				anc.Append(NewVarTuple[any](int64(m.Header.Level), int64(m.Header.Type), Bytes(m.Data)))
			}
		}
		return NewVarTuple[any](Bytes(buf[:n]), anc, int64(rflags), fromSockaddr(from))
	})
	RegisterNative("_socket.gethostname", func(args []any, kw []KW) any {
		h, err := os.Hostname()
		if err != nil {
			panic(newError("OSError", err.Error()))
		}
		return h
	})
	RegisterNative("_socket.gethostbyname_ex", func(args []any, kw []KW) any {
		host := hostArg(args[0])
		ips, err := net.LookupIP(host)
		if err != nil {
			panic(gaiError(err, host))
		}
		addrs := NewList[any]()
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil {
				addrs.Append(v4.String())
			}
		}
		if addrs.Len() == 0 {
			panic(gaiErrorCode(-9, "Address family for hostname not supported"))
		}
		return NewVarTuple[any](host, NewList[any](), addrs)
	})
	RegisterNative("_socket.gethostbyaddr", func(args []any, kw []KW) any {
		host := hostArg(args[0])
		ip := net.ParseIP(host)
		if ip == nil {
			ips, err := net.LookupIP(host)
			if err != nil || len(ips) == 0 {
				panic(gaiError(err, host))
			}
			ip = ips[0]
		}
		names, err := net.LookupAddr(ip.String())
		if err != nil || len(names) == 0 {
			e := &PyError{Type: "socket_pm_herror", Args: []any{int64(1), "Unknown host"}}
			panic(e)
		}
		addrs := NewList[any](ip.String())
		return NewVarTuple[any](strings.TrimSuffix(names[0], "."), NewList[any](), addrs)
	})
	// _socket.getaddrinfo(host, port, family, type, proto, flags)
	RegisterNative("_socket.getaddrinfo", func(args []any, kw []KW) any {
		var host string
		hasHost := !IsNone(args[0])
		if hasHost {
			host = hostArg(args[0])
		}
		var port int
		switch p := Unbox(args[1]).(type) {
		case nil, NoneT:
			port = 0
		case string:
			if n, err := strconv.Atoi(p); err == nil {
				port = n
			} else if n := serviceNumber(p, ""); n >= 0 {
				port = n
			} else {
				panic(gaiErrorCode(-8, "Servname not supported for ai_socktype"))
			}
		case Bytes:
			if n, err := strconv.Atoi(string(p)); err == nil {
				port = n
			} else {
				panic(gaiErrorCode(-8, "Servname not supported for ai_socktype"))
			}
		default:
			port = int(DynInt(p))
		}
		family, typ, proto := int(intArgAt(args, 2, 0)), int(intArgAt(args, 3, 0)), int(intArgAt(args, 4, 0))
		flags := int(intArgAt(args, 5, 0))
		const aiPassive, aiCanonname, aiNumericHost = 1, 2, 4
		var ips []net.IP
		switch {
		case !hasHost || host == "":
			if flags&aiPassive != 0 {
				ips = []net.IP{net.IPv4zero, net.IPv6zero}
			} else {
				ips = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
			}
		default:
			if ip := net.ParseIP(host); ip != nil {
				ips = []net.IP{ip}
			} else if flags&aiNumericHost != 0 {
				panic(gaiErrorCode(-2, "Name or service not known"))
			} else {
				var err error
				ips, err = net.LookupIP(host)
				if err != nil {
					panic(gaiError(err, host))
				}
			}
		}
		types := []int{syscall.SOCK_STREAM, syscall.SOCK_DGRAM, syscall.SOCK_RAW}
		if typ != 0 {
			types = []int{typ}
		}
		out := NewList[any]()
		for _, ip := range ips {
			v4 := ip.To4()
			fam := syscall.AF_INET6
			if v4 != nil {
				fam = syscall.AF_INET
			}
			if family != 0 && family != fam {
				continue
			}
			for _, t := range types {
				p := proto
				if p == 0 {
					switch t {
					case syscall.SOCK_STREAM:
						p = syscall.IPPROTO_TCP
					case syscall.SOCK_DGRAM:
						p = syscall.IPPROTO_UDP
					}
				}
				var sa any
				if fam == syscall.AF_INET {
					sa = NewVarTuple[any](v4.String(), int64(port))
				} else {
					sa = NewVarTuple[any](ip.String(), int64(port), int64(0), int64(0))
				}
				canon := ""
				if flags&aiCanonname != 0 {
					canon = host
				}
				out.Append(NewVarTuple[any](int64(fam), int64(t), int64(p), canon, sa))
			}
		}
		if out.Len() == 0 {
			panic(gaiErrorCode(-9, "Address family for hostname not supported"))
		}
		return out
	})
	RegisterNative("_socket.getservbyname", func(args []any, kw []KW) any {
		name := Str(args[0])
		proto := "tcp"
		if len(args) > 1 && !IsNone(args[1]) {
			proto = Str(args[1])
		}
		n := serviceNumber(name, proto)
		if n < 0 {
			panic(newError("OSError", "service/proto not found"))
		}
		return int64(n)
	})
	RegisterNative("_socket.getservbyport", func(args []any, kw []KW) any {
		port := int(DynInt(args[0]))
		proto := "tcp"
		if len(args) > 1 && !IsNone(args[1]) {
			proto = Str(args[1])
		}
		for name, n := range knownServices {
			if n == port && name != "" {
				_ = proto
				return name
			}
		}
		panic(newError("OSError", "port/proto not found"))
	})
	RegisterNative("_socket.getprotobyname", func(args []any, kw []KW) any {
		switch Str(args[0]) {
		case "tcp":
			return int64(6)
		case "udp":
			return int64(17)
		case "icmp":
			return int64(1)
		case "ip":
			return int64(0)
		case "ipv6":
			return int64(41)
		}
		panic(newError("OSError", "protocol not found"))
	})
	RegisterNative("_socket.inet_aton", func(args []any, kw []KW) any {
		s := hostArg(args[0])
		parts := strings.Split(s, ".")
		var vals []uint64
		for _, p := range parts {
			n, err := strconv.ParseUint(p, 0, 32)
			if err != nil || p == "" {
				panic(newError("OSError", "illegal IP address string passed to inet_aton"))
			}
			vals = append(vals, n)
		}
		var ip uint32
		switch len(vals) {
		case 1:
			ip = uint32(vals[0])
		case 2:
			ip = uint32(vals[0])<<24 | uint32(vals[1])
		case 3:
			ip = uint32(vals[0])<<24 | uint32(vals[1])<<16 | uint32(vals[2])
		case 4:
			ip = uint32(vals[0])<<24 | uint32(vals[1])<<16 | uint32(vals[2])<<8 | uint32(vals[3])
		default:
			panic(newError("OSError", "illegal IP address string passed to inet_aton"))
		}
		return Bytes([]byte{byte(ip >> 24), byte(ip >> 16), byte(ip >> 8), byte(ip)})
	})
	RegisterNative("_socket.inet_pton", func(args []any, kw []KW) any {
		fam := int(DynInt(args[0]))
		s := hostArg(args[1])
		ip := net.ParseIP(s)
		switch fam {
		case syscall.AF_INET:
			if ip == nil || ip.To4() == nil || strings.Contains(s, ":") {
				panic(newError("OSError", "illegal IP address string passed to inet_pton"))
			}
			return Bytes(ip.To4())
		case syscall.AF_INET6:
			if ip == nil || !strings.Contains(s, ":") {
				panic(newError("OSError", "illegal IP address string passed to inet_pton"))
			}
			return Bytes(ip.To16())
		}
		panic(sysError(syscall.EAFNOSUPPORT))
	})
	RegisterNative("_socket.inet_ntop", func(args []any, kw []KW) any {
		fam := int(DynInt(args[0]))
		b := bytesArg(args[1])
		switch fam {
		case syscall.AF_INET:
			if len(b) != 4 {
				panic(valueError("invalid length of packed IP address string"))
			}
			return net.IP(b).String()
		case syscall.AF_INET6:
			if len(b) != 16 {
				panic(valueError("invalid length of packed IP address string"))
			}
			return net.IP(b).String()
		}
		panic(valueError("unknown address family " + Str(int64(fam))))
	})
	RegisterNative("_socket.if_nameindex", func(args []any, kw []KW) any {
		out := NewList[any]()
		ifs, _ := net.Interfaces()
		for _, i := range ifs {
			out.Append(NewVarTuple[any](int64(i.Index), i.Name))
		}
		return out
	})
	RegisterNative("_socket.if_nametoindex", func(args []any, kw []KW) any {
		i, err := net.InterfaceByName(Str(args[0]))
		if err != nil {
			panic(newError("OSError", "no interface with this name"))
		}
		return int64(i.Index)
	})
	RegisterNative("_socket.if_indextoname", func(args []any, kw []KW) any {
		i, err := net.InterfaceByIndex(int(DynInt(args[0])))
		if err != nil {
			panic(newError("OSError", "no interface with this index"))
		}
		return i.Name
	})
}

var knownServices = map[string]int{"echo": 7, "ftp-data": 20, "ftp": 21, "ssh": 22, "telnet": 23, "smtp": 25, "domain": 53,
	"http": 80, "pop3": 110, "ntp": 123, "imap": 143, "snmp": 161, "https": 443, "smtps": 465, "submission": 587,
	"imaps": 993, "pop3s": 995, "www": 80}

func serviceNumber(name, proto string) int {
	if n, ok := knownServices[name]; ok {
		return n
	}
	return -1
}

func controlMessage(level, typ int, data []byte) []byte {
	b := make([]byte, syscall.CmsgSpace(len(data)))
	h := (*syscall.Cmsghdr)(unsafePointer(&b[0]))
	h.Level = int32(level)
	h.Type = int32(typ)
	h.SetLen(syscall.CmsgLen(len(data)))
	copy(b[syscall.CmsgLen(0):], data)
	return b
}
