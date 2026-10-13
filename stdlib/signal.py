"""signal: handlers for asynchronous events."""

from enum import IntEnum


class Signals(IntEnum):
    SIGHUP = 1
    SIGINT = 2
    SIGQUIT = 3
    SIGILL = 4
    SIGTRAP = 5
    SIGABRT = 6
    SIGBUS = 7
    SIGFPE = 8
    SIGKILL = 9
    SIGUSR1 = 10
    SIGSEGV = 11
    SIGUSR2 = 12
    SIGPIPE = 13
    SIGALRM = 14
    SIGTERM = 15
    SIGSTKFLT = 16
    SIGCHLD = 17
    SIGCONT = 18
    SIGSTOP = 19
    SIGTSTP = 20
    SIGTTIN = 21
    SIGTTOU = 22
    SIGURG = 23
    SIGXCPU = 24
    SIGXFSZ = 25
    SIGVTALRM = 26
    SIGPROF = 27
    SIGWINCH = 28
    SIGIO = 29
    SIGPWR = 30
    SIGSYS = 31
    SIGRTMIN = 34
    SIGRTMAX = 64


class Handlers(IntEnum):
    SIG_DFL = 0
    SIG_IGN = 1


SIG_DFL = Handlers.SIG_DFL
SIG_IGN = Handlers.SIG_IGN
NSIG = 65

SIGHUP = Signals.SIGHUP
SIGINT = Signals.SIGINT
SIGQUIT = Signals.SIGQUIT
SIGILL = Signals.SIGILL
SIGTRAP = Signals.SIGTRAP
SIGABRT = Signals.SIGABRT
SIGIOT = Signals.SIGABRT
SIGBUS = Signals.SIGBUS
SIGFPE = Signals.SIGFPE
SIGKILL = Signals.SIGKILL
SIGUSR1 = Signals.SIGUSR1
SIGSEGV = Signals.SIGSEGV
SIGUSR2 = Signals.SIGUSR2
SIGPIPE = Signals.SIGPIPE
SIGALRM = Signals.SIGALRM
SIGTERM = Signals.SIGTERM
SIGSTKFLT = Signals.SIGSTKFLT
SIGCHLD = Signals.SIGCHLD
SIGCONT = Signals.SIGCONT
SIGSTOP = Signals.SIGSTOP
SIGTSTP = Signals.SIGTSTP
SIGTTIN = Signals.SIGTTIN
SIGTTOU = Signals.SIGTTOU
SIGURG = Signals.SIGURG
SIGXCPU = Signals.SIGXCPU
SIGXFSZ = Signals.SIGXFSZ
SIGVTALRM = Signals.SIGVTALRM
SIGPROF = Signals.SIGPROF
SIGWINCH = Signals.SIGWINCH
SIGIO = Signals.SIGIO
SIGPOLL = Signals.SIGIO
SIGPWR = Signals.SIGPWR
SIGSYS = Signals.SIGSYS
SIGRTMIN = Signals.SIGRTMIN
SIGRTMAX = Signals.SIGRTMAX

ITIMER_REAL = 0
ITIMER_VIRTUAL = 1
ITIMER_PROF = 2

_NAMES = {1: "Hangup", 2: "Interrupt", 3: "Quit", 4: "Illegal instruction", 5: "Trace/breakpoint trap", 6: "Aborted", 7: "Bus error", 8: "Floating point exception", 9: "Killed", 10: "User defined signal 1", 11: "Segmentation fault", 12: "User defined signal 2", 13: "Broken pipe", 14: "Alarm clock", 15: "Terminated", 16: "Stack fault", 17: "Child exited", 18: "Continued", 19: "Stopped (signal)", 20: "Stopped", 21: "Stopped (tty input)", 22: "Stopped (tty output)", 23: "Urgent I/O condition", 24: "CPU time limit exceeded", 25: "File size limit exceeded", 26: "Virtual timer expired", 27: "Profiling timer expired", 28: "Window changed", 29: "I/O possible", 30: "Power failure", 31: "Bad system call"}


def default_int_handler(signalnum, frame):
    raise KeyboardInterrupt


def _handler_in(handler):
    if isinstance(handler, Handlers):
        return int(handler)
    return handler


def _handler_out(prev):
    if isinstance(prev, str):
        return default_int_handler
    if isinstance(prev, int):
        return Handlers(prev)
    return prev


def signal(signalnum, handler):
    return _handler_out(_native("signal.signal", int(signalnum), _handler_in(handler)))


def getsignal(signalnum):
    return _handler_out(_native("signal.getsignal", int(signalnum)))


def raise_signal(signalnum):
    _native("signal.raise_signal", int(signalnum))


def alarm(seconds):
    return _native("signal.alarm", seconds)


def setitimer(which, seconds, interval=0.0):
    return _native("signal.setitimer", which, seconds, interval)


def getitimer(which):
    return (0.0, 0.0)


def pause():
    _native("signal.pause")


def strsignal(signalnum):
    return _NAMES.get(int(signalnum))


def valid_signals():
    out = set()
    for n in range(1, 65):
        if n != 32 and n != 33:
            if n == 9 or n == 19 or True:
                out.add(Signals(n) if n in (1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 34, 64) else n)
    return out


def siginterrupt(signalnum, flag):
    pass


def set_wakeup_fd(fd, *, warn_on_full_buffer=True):
    return -1


def pthread_kill(thread_id, signalnum):
    _native("signal.kill", _native("signal.getpid"), int(signalnum))


def pthread_sigmask(how, mask):
    return set()
