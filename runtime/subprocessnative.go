package runtime

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// The natives of the subprocess module: starting a program on descriptors
// the caller chooses, waiting for it, and talking to its pipes.

const (
	spInherit = -1
	spPipe    = -2
	spDevnull = -3
	spStdout  = -4
)

func init() {
	// subprocess.spawn(argv, executable, cwd, env, stdin, stdout, stderr, new_session)
	// returns (pid, stdin_fd, stdout_fd, stderr_fd) with -1 for the ends the
	// caller does not get.
	RegisterNative("subprocess.spawn", func(args []any, kw []KW) any {
		var argv []string
		for c := DynIter(args[0]).Iter(); c.Next(); {
			s := Str(c.Value())
			if strings.IndexByte(s, 0) >= 0 {
				panic(valueError("embedded null byte"))
			}
			argv = append(argv, s)
		}
		exe := argv[0]
		if !IsNone(args[1]) {
			exe = Str(args[1])
		}
		var env []string
		envGiven := !IsNone(args[3])
		pathVar := os.Getenv("PATH")
		if envGiven {
			pathVar = ""
			for c := DynIter(args[3]).Iter(); c.Next(); {
				k := Str(c.Value())
				v := Str(DynGetItem(args[3], k))
				env = append(env, k+"="+v)
				if k == "PATH" {
					pathVar = v
				}
			}
		} else {
			env = os.Environ()
		}
		cwd := ""
		if !IsNone(args[2]) {
			cwd = Str(args[2])
		}
		// the program: a name with a slash is used as it is, the others are
		// searched for in PATH
		prog := exe
		if !strings.Contains(exe, "/") {
			found := ""
			for _, dir := range strings.Split(pathVar, ":") {
				if dir == "" {
					dir = "."
				}
				cand := dir + "/" + exe
				if st, err := os.Stat(cand); err == nil && !st.IsDir() && st.Mode()&0o111 != 0 {
					found = cand
					break
				}
			}
			if found == "" {
				panic(sysError(syscall.ENOENT, exe))
			}
			prog = found
		}
		modes := [3]int{int(DynInt(args[4])), int(DynInt(args[5])), int(DynInt(args[6]))}
		var files [3]uintptr
		var parent [3]int64
		var closeAfter []int
		var closeOnErr []int
		for i := range modes {
			parent[i] = -1
		}
		fail := func(err error, name string) {
			for _, fd := range closeAfter {
				syscall.Close(fd)
			}
			for _, fd := range closeOnErr {
				syscall.Close(fd)
			}
			panic(sysError(err, name))
		}
		for i := 0; i < 3; i++ {
			switch m := modes[i]; {
			case m == spInherit:
				files[i] = uintptr(i)
			case m == spPipe:
				var p [2]int
				if err := syscall.Pipe(p[:]); err != nil {
					fail(err, "")
				}
				syscall.CloseOnExec(p[0])
				syscall.CloseOnExec(p[1])
				if i == 0 {
					files[i], parent[i] = uintptr(p[0]), int64(p[1])
					closeAfter = append(closeAfter, p[0])
					closeOnErr = append(closeOnErr, p[1])
				} else {
					files[i], parent[i] = uintptr(p[1]), int64(p[0])
					closeAfter = append(closeAfter, p[1])
					closeOnErr = append(closeOnErr, p[0])
				}
			case m == spDevnull:
				fd, err := syscall.Open("/dev/null", syscall.O_RDWR|syscall.O_CLOEXEC, 0)
				if err != nil {
					fail(err, "")
				}
				files[i] = uintptr(fd)
				closeAfter = append(closeAfter, fd)
			case m == spStdout && i == 2:
				files[i] = files[1]
			case m >= 0:
				files[i] = uintptr(m)
			default:
				panic(valueError("bad file descriptor argument"))
			}
		}
		attr := &syscall.ProcAttr{Dir: cwd, Env: env, Files: files[:]}
		if Truth(args[7]) {
			attr.Sys = &syscall.SysProcAttr{Setsid: true}
		}
		var pid int
		var err error
		Blocking(func() { pid, err = syscall.ForkExec(prog, argv, attr) })
		for _, fd := range closeAfter {
			syscall.Close(fd)
		}
		if err != nil {
			for _, fd := range closeOnErr {
				syscall.Close(fd)
			}
			name := exe
			if cwd != "" && err == syscall.ENOENT {
				if _, serr := os.Stat(cwd); serr != nil {
					name = cwd
				}
			}
			panic(sysError(err, name))
		}
		return NewVarTuple[any](int64(pid), parent[0], parent[1], parent[2])
	})
	// subprocess.wait(pid, nohang): the return code, negative for a signal;
	// None when nohang and the process still runs.
	RegisterNative("subprocess.wait", func(args []any, kw []KW) any {
		pid := int(DynInt(args[0]))
		opts := 0
		if Truth(args[1]) {
			opts = syscall.WNOHANG
		}
		var ws syscall.WaitStatus
		var wpid int
		var err error
		if opts != 0 {
			for {
				wpid, err = syscall.Wait4(pid, &ws, opts, nil)
				if err != syscall.EINTR {
					break
				}
			}
		} else {
			Blocking(func() {
				for {
					wpid, err = syscall.Wait4(pid, &ws, 0, nil)
					if err != syscall.EINTR {
						break
					}
				}
			})
		}
		if err == syscall.ECHILD {
			return int64(0)
		}
		if err != nil {
			panic(sysError(err))
		}
		if wpid == 0 {
			return nil
		}
		switch {
		case ws.Exited():
			return int64(ws.ExitStatus())
		case ws.Signaled():
			return int64(-int(ws.Signal()))
		}
		return int64(ws)
	})
	// subprocess.communicate(stdin_fd, input, stdout_fd, stderr_fd, timeout):
	// writes the input, reads both outputs to their end; (out, err, timed_out).
	RegisterNative("subprocess.communicate", func(args []any, kw []KW) any {
		inFD, outFD, errFD := int(DynInt(args[0])), int(DynInt(args[2])), int(DynInt(args[3]))
		var input []byte
		if !IsNone(args[1]) {
			input = bytesArg(args[1])
		}
		timeout := -1.0
		if !IsNone(args[4]) {
			timeout = mathFloat(args[4])
		}
		var wg sync.WaitGroup
		var outBuf, errBuf []byte
		var mu sync.Mutex
		readAll := func(fd int, dst *[]byte) {
			defer wg.Done()
			buf := make([]byte, 65536)
			for {
				n, err := syscall.Read(fd, buf)
				if err == syscall.EINTR {
					continue
				}
				if n > 0 {
					mu.Lock()
					*dst = append(*dst, buf[:n]...)
					mu.Unlock()
				}
				if err != nil || n == 0 {
					return
				}
			}
		}
		if outFD >= 0 {
			wg.Add(1)
			go readAll(outFD, &outBuf)
		}
		if errFD >= 0 {
			wg.Add(1)
			go readAll(errFD, &errBuf)
		}
		if inFD >= 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer syscall.Close(inFD)
				data := input
				for len(data) > 0 {
					n, err := syscall.Write(inFD, data)
					if err == syscall.EINTR {
						continue
					}
					if err != nil {
						return
					}
					data = data[n:]
				}
			}()
		}
		done := make(chan struct{})
		go func() { wg.Wait(); close(done) }()
		timedOut := false
		Blocking(func() {
			if timeout < 0 {
				<-done
				return
			}
			t := time.NewTimer(time.Duration(timeout * float64(time.Second)))
			defer t.Stop()
			select {
			case <-done:
			case <-t.C:
				timedOut = true
			}
		})
		mu.Lock()
		defer mu.Unlock()
		var out, er any
		if outFD >= 0 {
			out = Bytes(outBuf)
		}
		if errFD >= 0 {
			er = Bytes(errBuf)
		}
		return NewVarTuple[any](out, er, timedOut)
	})
	RegisterNative("subprocess.which", func(args []any, kw []KW) any {
		p, err := exec.LookPath(Str(args[0]))
		if err != nil {
			return nil
		}
		return p
	})
}
