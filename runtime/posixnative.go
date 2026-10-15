package runtime

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// The natives of the posix module: the system calls os is made of. They
// work on raw descriptors, so what open() and os.open() hand out are the
// same numbers, and report failures as CPython does.

// sysError is the OSError of a failed call, from the errno inside err.
func sysError(err error, names ...string) *PyError {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return newError("OSError", err.Error())
	}
	text := errno.Error()
	text = strings.ToUpper(text[:1]) + text[1:]
	cls := "OSError"
	if c, ok := errnoClasses[int64(errno)]; ok {
		cls = c
	}
	e := &PyError{Type: cls, Args: []any{int64(errno), text}}
	e.Attrs = map[string]any{"errno": int64(errno), "strerror": text, "filename": nil, "filename2": nil}
	if len(names) > 0 && names[0] != "" {
		e.Attrs["filename"] = names[0]
	}
	if len(names) > 1 && names[1] != "" {
		e.Attrs["filename2"] = names[1]
	}
	return e
}

// pathArg is a path argument: str or bytes.
func pathArg(args []any, i int) string {
	if i >= len(args) {
		panic(newError("TypeError", "missing path argument"))
	}
	switch v := Unbox(args[i]).(type) {
	case string:
		if strings.IndexByte(v, 0) >= 0 {
			panic(valueError("embedded null byte"))
		}
		return v
	case Bytes:
		return string(v)
	}
	panic(newError("TypeError", "expected str, bytes or os.PathLike object, not "+TypeName(Unbox(args[i]))))
}

func intArgAt(args []any, i int, def int64) int64 {
	if i >= len(args) || IsNone(args[i]) {
		return def
	}
	return DynInt(args[i])
}

func init() {
	RegisterNative("posix.open", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		flags := intArgAt(args, 1, 0)
		mode := intArgAt(args, 2, 0o777)
		fd, err := syscall.Open(path, int(flags)|syscall.O_CLOEXEC, uint32(mode))
		if err != nil {
			panic(sysError(err, path))
		}
		return int64(fd)
	})
	RegisterNative("posix.close", func(args []any, kw []KW) any {
		if err := syscall.Close(int(DynInt(args[0]))); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("posix.read", func(args []any, kw []KW) any {
		fd, n := int(DynInt(args[0])), int(DynInt(args[1]))
		if n < 0 {
			panic(valueError("negative buffersize in read"))
		}
		buf := make([]byte, n)
		got, err := rawRead(fd, buf)
		if err != nil {
			panic(sysError(err))
		}
		return Bytes(buf[:got])
	})
	RegisterNative("posix.write", func(args []any, kw []KW) any {
		fd := int(DynInt(args[0]))
		data := bytesArg(args[1])
		n, err := rawWrite(fd, data)
		if err != nil {
			panic(sysError(err))
		}
		return int64(n)
	})
	RegisterNative("posix.lseek", func(args []any, kw []KW) any {
		off, err := syscall.Seek(int(DynInt(args[0])), DynInt(args[1]), int(DynInt(args[2])))
		if err != nil {
			panic(sysError(err))
		}
		return off
	})
	RegisterNative("posix.fstat", func(args []any, kw []KW) any {
		var st syscall.Stat_t
		if err := syscall.Fstat(int(DynInt(args[0])), &st); err != nil {
			panic(sysError(err))
		}
		return NewVarTuple(statFields(&st)...)
	})
	RegisterNative("posix.stat", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		var st syscall.Stat_t
		var err error
		if len(args) > 1 && !Truth(args[1]) {
			err = syscall.Lstat(path, &st)
		} else {
			err = syscall.Stat(path, &st)
		}
		if err != nil {
			panic(sysError(err, path))
		}
		return NewVarTuple(statFields(&st)...)
	})
	RegisterNative("posix.ftruncate", func(args []any, kw []KW) any {
		if err := syscall.Ftruncate(int(DynInt(args[0])), DynInt(args[1])); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("posix.truncate", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Truncate(path, DynInt(args[1])); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.fsync", func(args []any, kw []KW) any {
		if err := syscall.Fsync(int(DynInt(args[0]))); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("posix.dup", func(args []any, kw []KW) any {
		fd, err := syscall.Dup(int(DynInt(args[0])))
		if err != nil {
			panic(sysError(err))
		}
		syscall.CloseOnExec(fd)
		return int64(fd)
	})
	RegisterNative("posix.dup2", func(args []any, kw []KW) any {
		if err := syscall.Dup2(int(DynInt(args[0])), int(DynInt(args[1]))); err != nil {
			panic(sysError(err))
		}
		return DynInt(args[1])
	})
	RegisterNative("posix.pipe", func(args []any, kw []KW) any {
		var p [2]int
		if err := syscall.Pipe(p[:]); err != nil {
			panic(sysError(err))
		}
		syscall.CloseOnExec(p[0])
		syscall.CloseOnExec(p[1])
		return NewVarTuple[any](int64(p[0]), int64(p[1]))
	})
	RegisterNative("posix.isatty", func(args []any, kw []KW) any {
		var t syscall.Termios
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(DynInt(args[0])), uintptr(ioctlGetTermios), uintptr(unsafe.Pointer(&t)))
		return e == 0
	})
	RegisterNative("posix.get_terminal_size", func(args []any, kw []KW) any {
		fd := uintptr(1)
		if len(args) > 0 {
			fd = uintptr(DynInt(args[0]))
		}
		var ws [4]uint16
		if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws))); e != 0 {
			panic(sysError(e))
		}
		return NewVarTuple[any](int64(ws[1]), int64(ws[0]))
	})
	RegisterNative("posix.listdir", func(args []any, kw []KW) any {
		dir := "."
		if len(args) > 0 && !IsNone(args[0]) {
			dir = pathArg(args, 0)
		}
		f, err := os.Open(dir)
		if err != nil {
			panic(sysError(err, dir))
		}
		defer f.Close()
		names, err := f.Readdirnames(-1)
		if err != nil {
			panic(sysError(err, dir))
		}
		out := NewList[any]()
		for _, n := range names {
			out.Append(n)
		}
		return out
	})
	// posix.scandir(path): (name, type, inode) for each entry, in directory
	// order; type is "d", "f", "l" or "?" when the file system does not say.
	RegisterNative("posix.scandir", func(args []any, kw []KW) any {
		dir := "."
		if len(args) > 0 && !IsNone(args[0]) {
			dir = pathArg(args, 0)
		}
		f, err := os.Open(dir)
		if err != nil {
			panic(sysError(err, dir))
		}
		defer f.Close()
		if st, err := f.Stat(); err == nil && !st.IsDir() {
			panic(sysError(syscall.ENOTDIR, dir))
		}
		ents, err := f.ReadDir(-1)
		if err != nil {
			panic(sysError(err, dir))
		}
		out := NewList[any]()
		for _, e := range ents {
			kind := "?"
			switch t := e.Type(); {
			case t.IsDir():
				kind = "d"
			case t&os.ModeSymlink != 0:
				kind = "l"
			case t.IsRegular():
				kind = "f"
			}
			out.Append(NewVarTuple[any](e.Name(), kind))
		}
		return out
	})
	RegisterNative("posix.mkdir", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Mkdir(path, uint32(intArgAt(args, 1, 0o777))); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.rmdir", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Rmdir(path); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.unlink", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Unlink(path); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.rename", func(args []any, kw []KW) any {
		a, b := pathArg(args, 0), pathArg(args, 1)
		if err := syscall.Rename(a, b); err != nil {
			panic(sysError(err, a, b))
		}
		return nil
	})
	RegisterNative("posix.link", func(args []any, kw []KW) any {
		a, b := pathArg(args, 0), pathArg(args, 1)
		if err := syscall.Link(a, b); err != nil {
			panic(sysError(err, a, b))
		}
		return nil
	})
	RegisterNative("posix.symlink", func(args []any, kw []KW) any {
		a, b := pathArg(args, 0), pathArg(args, 1)
		if err := syscall.Symlink(a, b); err != nil {
			panic(sysError(err, a, b))
		}
		return nil
	})
	RegisterNative("posix.readlink", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		buf := make([]byte, 4096)
		n, err := syscall.Readlink(path, buf)
		if err != nil {
			panic(sysError(err, path))
		}
		return string(buf[:n])
	})
	RegisterNative("posix.chmod", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Chmod(path, uint32(DynInt(args[1]))); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.fchmod", func(args []any, kw []KW) any {
		if err := syscall.Fchmod(int(DynInt(args[0])), uint32(DynInt(args[1]))); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("posix.chown", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Chown(path, int(DynInt(args[1])), int(DynInt(args[2]))); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.lchown", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Lchown(path, int(DynInt(args[1])), int(DynInt(args[2]))); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.utime", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		ts := []syscall.Timespec{syscall.NsecToTimespec(DynInt(args[1])), syscall.NsecToTimespec(DynInt(args[2]))}
		if err := syscall.UtimesNano(path, ts); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.access", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		return syscall.Access(path, uint32(DynInt(args[1]))) == nil
	})
	RegisterNative("posix.chdir", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Chdir(path); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.getcwd", func(args []any, kw []KW) any {
		wd, err := syscall.Getwd()
		if err != nil {
			panic(sysError(err))
		}
		return wd
	})
	RegisterNative("posix.umask", func(args []any, kw []KW) any {
		return int64(syscall.Umask(int(DynInt(args[0]))))
	})
	RegisterNative("posix.mkfifo", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		if err := syscall.Mkfifo(path, uint32(intArgAt(args, 1, 0o666))); err != nil {
			panic(sysError(err, path))
		}
		return nil
	})
	RegisterNative("posix.ids", func(args []any, kw []KW) any {
		return NewVarTuple[any](int64(syscall.Getuid()), int64(syscall.Getgid()), int64(syscall.Geteuid()), int64(syscall.Getegid()))
	})
	RegisterNative("posix.uname", func(args []any, kw []KW) any {
		u := unameFields()
		return NewVarTuple[any](u[0], u[1], u[2], u[3], u[4])
	})
	RegisterNative("posix.strerror", func(args []any, kw []KW) any {
		n := DynInt(args[0])
		if n == 0 {
			return "Success"
		}
		text := syscall.Errno(n).Error()
		if strings.HasPrefix(text, "errno ") {
			return "Unknown error " + Str(n)
		}
		return strings.ToUpper(text[:1]) + text[1:]
	})
	RegisterNative("posix.statvfs", func(args []any, kw []KW) any {
		path := pathArg(args, 0)
		var st syscall.Statfs_t
		if err := syscall.Statfs(path, &st); err != nil {
			panic(sysError(err, path))
		}
		return NewVarTuple[any](int64(st.Bsize), int64(st.Bsize), int64(st.Blocks), int64(st.Bfree), int64(st.Bavail), int64(st.Files), int64(st.Ffree), int64(st.Ffree), int64(0), int64(255))
	})
	RegisterNative("posix.system", func(args []any, kw []KW) any {
		cmd := exec.Command("/bin/sh", "-c", Str(args[0]))
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		err := cmd.Run()
		if err == nil {
			return int64(0)
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return int64(ee.ExitCode() << 8)
		}
		return int64(127 << 8)
	})
	// posix.home(name): the home directory of a user (None for the current
	// user's HOME fallback), from the password file.
	RegisterNative("posix.home", func(args []any, kw []KW) any {
		name := ""
		if len(args) > 0 && !IsNone(args[0]) {
			name = Str(args[0])
		}
		data, err := os.ReadFile("/etc/passwd")
		if err != nil {
			return nil
		}
		uid := Str(int64(syscall.Getuid()))
		for _, line := range strings.Split(string(data), "\n") {
			f := strings.Split(line, ":")
			if len(f) < 6 {
				continue
			}
			if (name != "" && f[0] == name) || (name == "" && f[2] == uid) {
				return f[5]
			}
		}
		return nil
	})
	// posix.passwd() and posix.group(): the lines of the user and group
	// databases, split into fields.
	RegisterNative("posix.passwd", func(args []any, kw []KW) any {
		return dbLines("/etc/passwd", 7)
	})
	RegisterNative("posix.group", func(args []any, kw []KW) any {
		return dbLines("/etc/group", 4)
	})
	RegisterNative("posix.getloadavg", func(args []any, kw []KW) any {
		data, err := os.ReadFile("/proc/loadavg")
		if err != nil {
			panic(newError("OSError", "Load averages are unobtainable"))
		}
		f := strings.Fields(string(data))
		return NewVarTuple[any](parseFloat(f[0]), parseFloat(f[1]), parseFloat(f[2]))
	})
	RegisterNative("posix.getlogin", func(args []any, kw []KW) any {
		if u := os.Getenv("LOGNAME"); u != "" {
			return u
		}
		panic(sysError(syscall.ENOTTY))
	})
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func dbLines(path string, n int) any {
	out := NewList[any]()
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" || line[0] == '#' {
			continue
		}
		f := strings.Split(line, ":")
		for len(f) < n {
			f = append(f, "")
		}
		row := make([]any, 0, n)
		for _, x := range f[:n] {
			row = append(row, x)
		}
		out.Append(NewVarTuple(row...))
	}
	return out
}

func init() {
	RegisterNative("resource.getrusage", func(args []any, kw []KW) any {
		who := int(DynInt(args[0]))
		var ru syscall.Rusage
		if err := syscall.Getrusage(who, &ru); err != nil {
			panic(sysError(err))
		}
		tv := func(t syscall.Timeval) float64 { return float64(t.Sec) + float64(t.Usec)/1e6 }
		return NewVarTuple[any](tv(ru.Utime), tv(ru.Stime), int64(ru.Maxrss), int64(ru.Ixrss), int64(ru.Idrss), int64(ru.Isrss),
			int64(ru.Minflt), int64(ru.Majflt), int64(ru.Nswap), int64(ru.Inblock), int64(ru.Oublock), int64(ru.Msgsnd),
			int64(ru.Msgrcv), int64(ru.Nsignals), int64(ru.Nvcsw), int64(ru.Nivcsw))
	})
	RegisterNative("resource.getrlimit", func(args []any, kw []KW) any {
		var rl syscall.Rlimit
		if err := syscall.Getrlimit(int(DynInt(args[0])), &rl); err != nil {
			panic(valueError("invalid resource specified"))
		}
		conv := func(v uint64) any {
			if v == ^uint64(0) {
				return int64(-1)
			}
			return int64(v)
		}
		return NewVarTuple[any](conv(rl.Cur), conv(rl.Max))
	})
	RegisterNative("resource.setrlimit", func(args []any, kw []KW) any {
		conv := func(v int64) uint64 {
			if v < 0 {
				return ^uint64(0)
			}
			return uint64(v)
		}
		rl := syscall.Rlimit{Cur: conv(DynInt(args[1])), Max: conv(DynInt(args[2]))}
		if err := syscall.Setrlimit(int(DynInt(args[0])), &rl); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("fcntl.flock", func(args []any, kw []KW) any {
		if err := syscall.Flock(int(DynInt(args[0])), int(DynInt(args[1]))); err != nil {
			panic(sysError(err))
		}
		return nil
	})
	RegisterNative("fcntl.fcntl", func(args []any, kw []KW) any {
		fd, cmd := DynInt(args[0]), DynInt(args[1])
		arg := int64(0)
		if len(args) > 2 {
			arg = DynInt(args[2])
		}
		r, _, e := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), uintptr(cmd), uintptr(arg))
		if e != 0 {
			panic(sysError(e))
		}
		return int64(r)
	})
	// select.select(rlist, wlist, xlist, timeout)
	RegisterNative("select.select", func(args []any, kw []KW) any {
		fdsOf := func(v any) []int {
			var out []int
			for c := DynIter(v).Iter(); c.Next(); {
				x := c.Value()
				if n, ok := intOf(Unbox(x)); ok {
					out = append(out, int(n))
				} else {
					out = append(out, int(DynInt(DynCallMethod(x, "fileno", nil, nil))))
				}
			}
			return out
		}
		lists := [3][]int{fdsOf(args[0]), fdsOf(args[1]), fdsOf(args[2])}
		var sets [3]syscall.FdSet
		maxfd := -1
		for i := range lists {
			for _, fd := range lists[i] {
				if fd < 0 || fd >= 1024 {
					panic(valueError("filedescriptor out of range in select()"))
				}
				sets[i].Bits[fd/64] |= 1 << (uint(fd) % 64)
				if fd > maxfd {
					maxfd = fd
				}
			}
		}
		var tv *syscall.Timeval
		if len(args) > 3 && !IsNone(args[3]) {
			secs := mathFloat(args[3])
			if secs < 0 {
				panic(valueError("timeout must be non-negative"))
			}
			t := syscall.NsecToTimeval(int64(secs * 1e9))
			tv = &t
		}
		var err error
		Blocking(func() {
			for {
				err = sysSelect(maxfd+1, &sets[0], &sets[1], &sets[2], tv)
				if err != syscall.EINTR {
					break
				}
			}
		})
		if err != nil {
			panic(sysError(err))
		}
		res := make([]any, 3)
		for i := range lists {
			out := NewList[any]()
			for _, fd := range lists[i] {
				if sets[i].Bits[fd/64]&(1<<(uint(fd)%64)) != 0 {
					out.Append(int64(fd))
				}
			}
			res[i] = out
		}
		return NewVarTuple(res...)
	})
}
