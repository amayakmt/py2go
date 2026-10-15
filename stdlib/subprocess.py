"""subprocess: spawn new processes, connect to their pipes and get their return codes."""
import os
import time as _time
import signal as _signal

PIPE = -1
STDOUT = -2
DEVNULL = -3

__all__ = ["Popen", "PIPE", "STDOUT", "call", "check_call", "getstatusoutput",
           "getoutput", "check_output", "run", "CalledProcessError", "DEVNULL",
           "SubprocessError", "TimeoutExpired", "CompletedProcess"]


class SubprocessError(Exception):
    pass


class CalledProcessError(SubprocessError):
    def __init__(self, returncode, cmd, output=None, stderr=None):
        self.returncode = returncode
        self.cmd = cmd
        self.output = output
        self.stderr = stderr

    def __str__(self):
        if self.returncode and self.returncode < 0:
            try:
                return "Command '%s' died with %r." % (self.cmd, _signal.Signals(-self.returncode))
            except ValueError:
                return "Command '%s' died with unknown signal %d." % (self.cmd, -self.returncode)
        else:
            return "Command '%s' returned non-zero exit status %d." % (self.cmd, self.returncode)

    @property
    def stdout(self):
        return self.output

    @stdout.setter
    def stdout(self, value):
        self.output = value


class TimeoutExpired(SubprocessError):
    def __init__(self, cmd, timeout, output=None, stderr=None):
        self.cmd = cmd
        self.timeout = timeout
        self.output = output
        self.stderr = stderr

    def __str__(self):
        return "Command '%s' timed out after %s seconds" % (self.cmd, self.timeout)

    @property
    def stdout(self):
        return self.output

    @stdout.setter
    def stdout(self, value):
        self.output = value


def list2cmdline(seq):
    result = []
    needquote = False
    for arg in map(os.fsdecode, seq):
        bs_buf = []
        if result:
            result.append(" ")
        needquote = (" " in arg) or ("\t" in arg) or not arg
        if needquote:
            result.append('"')
        for c in arg:
            if c == "\\":
                bs_buf.append(c)
            elif c == '"':
                result.append("\\" * len(bs_buf) * 2)
                bs_buf = []
                result.append('\\"')
            else:
                if bs_buf:
                    result.extend(bs_buf)
                    bs_buf = []
                result.append(c)
        if bs_buf:
            result.extend(bs_buf)
        if needquote:
            result.extend(bs_buf)
            result.append('"')
    return "".join(result)


class CompletedProcess:
    def __init__(self, args, returncode, stdout=None, stderr=None):
        self.args = args
        self.returncode = returncode
        self.stdout = stdout
        self.stderr = stderr

    def __repr__(self):
        args = ["args={!r}".format(self.args), "returncode={!r}".format(self.returncode)]
        if self.stdout is not None:
            args.append("stdout={!r}".format(self.stdout))
        if self.stderr is not None:
            args.append("stderr={!r}".format(self.stderr))
        return "{}({})".format(type(self).__name__, ", ".join(args))

    def check_returncode(self):
        if self.returncode:
            raise CalledProcessError(self.returncode, self.args, self.stdout, self.stderr)


def _translate_newlines(data, encoding, errors):
    data = data.decode(encoding, errors)
    return data.replace("\r\n", "\n").replace("\r", "\n")


class Popen:
    def __init__(self, args, bufsize=-1, executable=None, stdin=None, stdout=None, stderr=None,
                 preexec_fn=None, close_fds=True, shell=False, cwd=None, env=None,
                 universal_newlines=None, startupinfo=None, creationflags=0, restore_signals=True,
                 start_new_session=False, pass_fds=(), *, user=None, group=None, extra_groups=None,
                 encoding=None, errors=None, text=None, umask=-1, pipesize=-1, process_group=None):
        if bufsize is None:
            bufsize = -1
        if not isinstance(bufsize, int):
            raise TypeError("bufsize must be an integer")
        if preexec_fn is not None:
            raise ValueError("preexec_fn is not supported")
        self.args = args
        self.stdin = None
        self.stdout = None
        self.stderr = None
        self.pid = None
        self.returncode = None
        self.encoding = encoding
        self.errors = errors
        self.universal_newlines = universal_newlines
        self._waitpid_lock = None
        self._communication_started = False
        self.text_mode = encoding or errors or text or universal_newlines
        if self.text_mode and encoding is None:
            self.encoding = "utf-8"
        if text is not None and universal_newlines is not None and bool(universal_newlines) != bool(text):
            raise SubprocessError("Cannot disambiguate when both text and universal_newlines are supplied but different. Pass one or the other.")
        self._closed_child_pipe_fds = False

        if isinstance(args, (str, bytes)):
            argv = [args]
        elif isinstance(args, os.PathLike):
            if shell:
                raise TypeError("path-like args is not allowed when shell is true")
            argv = [args]
        else:
            argv = list(args)
        argv = [os.fsdecode(a) for a in argv]
        if shell:
            unix_shell = "/bin/sh"
            argv = [unix_shell, "-c"] + argv
            if executable:
                argv[0] = executable
        if not argv:
            raise IndexError("list index out of range")
        if executable is not None:
            executable = os.fsdecode(executable)
        elif shell:
            executable = "/bin/sh"
        if cwd is not None:
            cwd = os.fsdecode(cwd)
        envmap = None
        if env is not None:
            envmap = {}
            for k in env:
                envmap[os.fsdecode(k)] = os.fsdecode(env[k])

        in_mode = self._fd_mode(stdin, 0)
        out_mode = self._fd_mode(stdout, 1)
        err_mode = self._fd_mode(stderr, 2)
        pid, p2cwrite, c2pread, errread = _native("subprocess.spawn", argv, executable, cwd, envmap,
                                                  in_mode, out_mode, err_mode, start_new_session)
        self.pid = pid
        if p2cwrite >= 0:
            self.stdin = self._open_pipe(p2cwrite, "w", bufsize)
        if c2pread >= 0:
            self.stdout = self._open_pipe(c2pread, "r", bufsize)
        if errread >= 0:
            self.stderr = self._open_pipe(errread, "r", bufsize)
        self._stdin_fd = p2cwrite
        self._stdout_fd = c2pread
        self._stderr_fd = errread

    def _fd_mode(self, spec, index):
        if spec is None:
            return -1
        if spec == PIPE:
            return -2
        if spec == DEVNULL:
            return -3
        if spec == STDOUT:
            if index != 2:
                raise ValueError("STDOUT can only be used for stderr")
            return -4
        if isinstance(spec, int):
            return spec
        return spec.fileno()

    def _open_pipe(self, fd, direction, bufsize):
        if direction == "w":
            if self.text_mode:
                return open(fd, "w", bufsize, self.encoding, self.errors)
            return open(fd, "wb", bufsize)
        if self.text_mode:
            return open(fd, "r", bufsize, self.encoding, self.errors)
        return open(fd, "rb", bufsize)

    def __repr__(self):
        obj_repr = "<{}: returncode: {} args: {!r}>".format(self.__class__.__name__, self.returncode, self.args)
        if len(obj_repr) > 80:
            obj_repr = obj_repr[:76] + "...>"
        return obj_repr

    def __enter__(self):
        return self

    def __exit__(self, exc_type, value, traceback):
        if self.stdout:
            self.stdout.close()
        if self.stderr:
            self.stderr.close()
        try:
            if self.stdin:
                self.stdin.close()
        finally:
            if exc_type == KeyboardInterrupt:
                return
            self.wait()

    def poll(self):
        if self.returncode is None:
            rc = _native("subprocess.wait", self.pid, True)
            if rc is not None:
                self.returncode = rc
        return self.returncode

    def wait(self, timeout=None):
        if self.returncode is not None:
            return self.returncode
        if timeout is None:
            self.returncode = _native("subprocess.wait", self.pid, False)
            return self.returncode
        endtime = _time.monotonic() + timeout
        delay = 0.0005
        while True:
            if self.poll() is not None:
                return self.returncode
            remaining = endtime - _time.monotonic()
            if remaining <= 0:
                raise TimeoutExpired(self.args, timeout)
            delay = min(delay * 2, remaining, 0.05)
            _time.sleep(delay)

    def communicate(self, input=None, timeout=None):
        if self._communication_started and input:
            raise ValueError("Cannot send input after starting communication")
        self._communication_started = True
        if self.text_mode and input is not None and isinstance(input, str):
            input = input.encode(self.encoding, self.errors or "strict")
        elif input is not None and isinstance(input, str):
            raise TypeError("a bytes-like object is required, not 'str'")
        in_fd = -1
        if self.stdin is not None and input is not None:
            self.stdin.flush()
            in_fd = self._stdin_fd
        elif self.stdin is not None:
            try:
                self.stdin.close()
            except BrokenPipeError:
                pass
        out_fd = self._stdout_fd if self.stdout is not None else -1
        err_fd = self._stderr_fd if self.stderr is not None else -1
        out = None
        err = None
        if in_fd >= 0 or out_fd >= 0 or err_fd >= 0:
            out, err, timed_out = _native("subprocess.communicate", in_fd, input, out_fd, err_fd, timeout)
            if timed_out:
                raise TimeoutExpired(self.args, timeout, out, err)
        if self.stdin is not None and in_fd >= 0:
            try:
                self.stdin.close()
            except OSError:
                pass
        if self.stdout is not None:
            self.stdout.close()
        if self.stderr is not None:
            self.stderr.close()
        if timeout is None:
            self.wait()
        else:
            self.wait(timeout)
        if self.text_mode:
            if out is not None:
                out = _translate_newlines(out, self.encoding, self.errors or "strict")
            if err is not None:
                err = _translate_newlines(err, self.encoding, self.errors or "strict")
        return (out, err)

    def send_signal(self, sig):
        self.poll()
        if self.returncode is not None:
            return
        try:
            os.kill(self.pid, sig)
        except ProcessLookupError:
            pass

    def terminate(self):
        self.send_signal(_signal.SIGTERM)

    def kill(self):
        self.send_signal(_signal.SIGKILL)


def call(*popenargs, timeout=None, **kwargs):
    with Popen(*popenargs, **kwargs) as p:
        try:
            return p.wait(timeout=timeout)
        except:
            p.kill()
            raise


def check_call(*popenargs, **kwargs):
    retcode = call(*popenargs, **kwargs)
    if retcode:
        cmd = kwargs.get("args")
        if cmd is None:
            cmd = popenargs[0]
        raise CalledProcessError(retcode, cmd)
    return 0


def check_output(*popenargs, timeout=None, **kwargs):
    for kw in ("stdout", "check"):
        if kw in kwargs:
            raise ValueError(kw + " argument not allowed, it will be overridden.")
    if "input" in kwargs and kwargs["input"] is None:
        if kwargs.get("universal_newlines") or kwargs.get("text") or kwargs.get("encoding") or kwargs.get("errors"):
            empty = ""
        else:
            empty = b""
        kwargs["input"] = empty
    return run(*popenargs, stdout=PIPE, timeout=timeout, check=True, **kwargs).stdout


def run(*popenargs, input=None, capture_output=False, timeout=None, check=False, **kwargs):
    if input is not None:
        if kwargs.get("stdin") is not None:
            raise ValueError("stdin and input arguments may not both be used.")
        kwargs["stdin"] = PIPE
    if capture_output:
        if kwargs.get("stdout") is not None or kwargs.get("stderr") is not None:
            raise ValueError("stdout and stderr arguments may not be used with capture_output.")
        kwargs["stdout"] = PIPE
        kwargs["stderr"] = PIPE
    with Popen(*popenargs, **kwargs) as process:
        try:
            stdout, stderr = process.communicate(input, timeout=timeout)
        except TimeoutExpired as exc:
            process.kill()
            process.wait()
            raise
        except:
            process.kill()
            raise
        retcode = process.poll()
        pargs = process.args
        if check and retcode:
            raise CalledProcessError(retcode, pargs, output=stdout, stderr=stderr)
    return CompletedProcess(pargs, retcode, stdout, stderr)


def getstatusoutput(cmd, *, encoding=None, errors=None):
    try:
        data = check_output(cmd, shell=True, text=True, stderr=STDOUT, encoding=encoding, errors=errors)
        exitcode = 0
    except CalledProcessError as ex:
        data = ex.output
        exitcode = ex.returncode
    if data[-1:] == "\n":
        data = data[:-1]
    return exitcode, data


def getoutput(cmd, *, encoding=None, errors=None):
    return getstatusoutput(cmd, encoding=encoding, errors=errors)[1]
