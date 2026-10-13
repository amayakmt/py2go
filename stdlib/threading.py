"""threading: threads, locks and the objects built on them."""

import sys as _sys
import _thread
import functools
from time import monotonic as _time
from collections import deque as _deque

from builtins import threading_pm_BrokenBarrierError as BrokenBarrierError
from _thread import get_ident, get_native_id, allocate_lock as _allocate_lock, TIMEOUT_MAX

ThreadError = RuntimeError
_start_new_thread = _thread.start_new_thread


def Lock():
    return _thread.allocate_lock()


def RLock():
    return _thread.RLock()


def local():
    return _thread._local()


def stack_size(size=0):
    return _thread.stack_size(size)


def gettrace():
    return None


def getprofile():
    return None


def settrace(func):
    pass


def setprofile(func):
    pass


def settrace_all_threads(func):
    pass


def setprofile_all_threads(func):
    pass


class _State:
    def __init__(self):
        self.counter = 0
        self.active = {}
        self.limbo = {}
        self.dangling = []
        self.main = None
        self.excepthook = None
        self.atexits = []


_st = _State()


def _newname(template):
    _st.counter += 1
    return template + str(_st.counter)


class Condition:
    def __init__(self, lock=None):
        if lock is None:
            lock = RLock()
        self._lock = lock
        self._waiters = _deque()
        self._has_owned = hasattr(lock, "_is_owned")
        self._has_save = hasattr(lock, "_release_save")

    def __enter__(self):
        return self._lock.__enter__()

    def __exit__(self, t, v, tb):
        return self._lock.__exit__(t, v, tb)

    def __repr__(self):
        return "<Condition(" + repr(self._lock) + ", " + str(len(self._waiters)) + ")>"

    def acquire(self, *args):
        return self._lock.acquire(*args)

    def release(self):
        return self._lock.release()

    def _release_save(self):
        if self._has_save:
            return self._lock._release_save()
        self._lock.release()
        return None

    def _acquire_restore(self, x):
        if self._has_save:
            self._lock._acquire_restore(x)
        else:
            self._lock.acquire()

    def _is_owned(self):
        if self._has_owned:
            return self._lock._is_owned()
        if self._lock.acquire(False):
            self._lock.release()
            return False
        return True

    def wait(self, timeout=None):
        if not self._is_owned():
            raise RuntimeError("cannot wait on un-acquired lock")
        waiter = _allocate_lock()
        waiter.acquire()
        self._waiters.append(waiter)
        saved_state = self._release_save()
        gotit = False
        try:
            if timeout is None:
                waiter.acquire()
                gotit = True
            else:
                if timeout > 0:
                    gotit = waiter.acquire(True, timeout)
                else:
                    gotit = waiter.acquire(False)
            return gotit
        finally:
            self._acquire_restore(saved_state)
            if not gotit:
                try:
                    self._waiters.remove(waiter)
                except ValueError:
                    pass

    def wait_for(self, predicate, timeout=None):
        endtime = None
        waittime = timeout
        result = predicate()
        while not result:
            if waittime is not None:
                if endtime is None:
                    endtime = _time() + waittime
                else:
                    waittime = endtime - _time()
                    if waittime <= 0:
                        break
            self.wait(waittime)
            result = predicate()
        return result

    def notify(self, n=1):
        if not self._is_owned():
            raise RuntimeError("cannot notify on un-acquired lock")
        waiters = self._waiters
        while waiters and n > 0:
            waiter = waiters[0]
            try:
                waiter.release()
            except RuntimeError:
                pass
            else:
                n -= 1
            try:
                waiters.remove(waiter)
            except ValueError:
                pass

    def notify_all(self):
        self.notify(len(self._waiters))

    def notifyAll(self):
        self.notify(len(self._waiters))


class Semaphore:
    def __init__(self, value=1):
        if value < 0:
            raise ValueError("semaphore initial value must be >= 0")
        self._cond = Condition(Lock())
        self._value = value

    def __repr__(self):
        return "<" + type(self).__name__ + " value=" + str(self._value) + ">"

    def acquire(self, blocking=True, timeout=None):
        if not blocking and timeout is not None:
            raise ValueError("can't specify timeout for non-blocking acquire")
        rc = False
        endtime = None
        with self._cond:
            while self._value == 0:
                if not blocking:
                    break
                if timeout is not None:
                    if endtime is None:
                        endtime = _time() + timeout
                    else:
                        timeout = endtime - _time()
                        if timeout <= 0:
                            break
                self._cond.wait(timeout)
            else:
                self._value -= 1
                rc = True
        return rc

    def __enter__(self):
        return self.acquire()

    def release(self, n=1):
        if n < 1:
            raise ValueError("n must be one or more")
        with self._cond:
            self._value += n
            self._cond.notify(n)

    def __exit__(self, t, v, tb):
        self.release()


class BoundedSemaphore(Semaphore):
    def __init__(self, value=1):
        Semaphore.__init__(self, value)
        self._initial_value = value

    def release(self, n=1):
        if n < 1:
            raise ValueError("n must be one or more")
        with self._cond:
            if self._value + n > self._initial_value:
                raise ValueError("Semaphore released too many times")
            self._value += n
            self._cond.notify(n)


class Event:
    def __init__(self):
        self._cond = Condition(Lock())
        self._flag = False

    def is_set(self):
        return self._flag

    def isSet(self):
        return self._flag

    def set(self):
        with self._cond:
            self._flag = True
            self._cond.notify_all()

    def clear(self):
        with self._cond:
            self._flag = False

    def wait(self, timeout=None):
        with self._cond:
            signaled = self._flag
            if not signaled:
                signaled = self._cond.wait(timeout)
            return signaled


class Barrier:
    def __init__(self, parties, action=None, timeout=None):
        if parties < 1:
            raise ValueError("parties must be >= 1")
        self._cond = Condition(Lock())
        self._action = action
        self._timeout = timeout
        self._parties = parties
        self._state = 0
        self._count = 0

    def wait(self, timeout=None):
        if timeout is None:
            timeout = self._timeout
        with self._cond:
            self._enter()
            index = self._count
            self._count += 1
            try:
                if index + 1 == self._parties:
                    self._release()
                else:
                    self._wait(timeout)
                return index
            finally:
                self._count -= 1
                self._exit()

    def _enter(self):
        while self._state == 1 or self._state == -1:
            self._cond.wait()
        if self._state < 0:
            raise BrokenBarrierError

    def _release(self):
        try:
            if self._action:
                self._action()
            self._state = 1
            self._cond.notify_all()
        except:
            self._break()
            raise

    def _wait(self, timeout):
        if not self._cond.wait_for(lambda: self._state != 0, timeout):
            self._break()
            raise BrokenBarrierError
        if self._state < 0:
            raise BrokenBarrierError

    def _exit(self):
        if self._count == 0:
            if self._state == 1 or self._state == -1:
                self._state = 0
                self._cond.notify_all()

    def reset(self):
        with self._cond:
            if self._count > 0:
                if self._state == 0:
                    self._state = -1
                elif self._state == -2:
                    self._state = -1
            else:
                self._state = 0
            self._cond.notify_all()

    def abort(self):
        with self._cond:
            self._break()

    def _break(self):
        self._state = -2
        self._cond.notify_all()

    @property
    def parties(self):
        return self._parties

    @property
    def n_waiting(self):
        if self._state == 0:
            return self._count
        return 0

    @property
    def broken(self):
        return self._state == -2


class ExceptHookArgs:
    def __init__(self, exc_type, exc_value, exc_traceback, thread):
        self.exc_type = exc_type
        self.exc_value = exc_value
        self.exc_traceback = exc_traceback
        self.thread = thread


def _default_excepthook(args):
    if args.exc_type is SystemExit:
        return
    name = str(get_ident())
    if args.thread is not None:
        name = args.thread.name
    _sys.stderr.write("Exception in thread " + name + ":\n")
    _sys.stderr.write("Traceback (most recent call last):\n")
    _sys.stderr.write(args.exc_type.__name__ + ": " + str(args.exc_value) + "\n")


class _DefaultHook:
    def __call__(self, args):
        _default_excepthook(args)


excepthook = _DefaultHook()
__excepthook__ = excepthook


class Thread:
    def __init__(self, group=None, target=None, name=None, args=(), kwargs=None, *, daemon=None):
        if group is not None:
            raise AssertionError("group argument must be None for now")
        if kwargs is None:
            kwargs = {}
        if name is None:
            name = _newname("Thread-")
            if target is not None:
                if hasattr(target, "__name__"):
                    name = name + " (" + target.__name__ + ")"
        self._target = target
        self._name = str(name)
        self._args = args
        self._kwargs = kwargs
        if daemon is not None:
            self._daemonic = daemon
        else:
            self._daemonic = current_thread().daemon
        self._ident = None
        self._native_id = None
        self._started = Event()
        self._is_stopped = False
        self._initialized = True
        self._done_lock = _allocate_lock()

    def __repr__(self):
        status = "initial"
        if self._started.is_set():
            status = "started"
        if self._is_stopped:
            status = "stopped"
        if self._daemonic:
            status += " daemon"
        if self._ident is not None:
            status += " " + str(self._ident)
        return "<" + type(self).__name__ + "(" + self._name + ", " + status + ")>"

    def start(self):
        if not self._initialized:
            raise RuntimeError("thread.__init__() not called")
        if self._started.is_set():
            raise RuntimeError("threads can only be started once")
        _st.limbo[id(self)] = self
        self._done_lock.acquire()
        try:
            _start_new_thread(self._bootstrap, ())
        except BaseException:
            _st.limbo.pop(id(self), None)
            raise
        self._started.wait()

    def run(self):
        try:
            if self._target is not None:
                self._target(*self._args, **self._kwargs)
        finally:
            self._target = None
            self._args = ()
            self._kwargs = {}

    def _bootstrap(self):
        self._ident = get_ident()
        self._native_id = get_native_id()
        _st.active[self._ident] = self
        _st.limbo.pop(id(self), None)
        self._started.set()
        try:
            try:
                self.run()
            except SystemExit:
                pass
            except BaseException as exc:
                excepthook(ExceptHookArgs(type(exc), exc, None, self))
        finally:
            self._is_stopped = True
            _st.active.pop(self._ident, None)
            self._done_lock.release()

    def join(self, timeout=None):
        if not self._initialized:
            raise RuntimeError("Thread.__init__() not called")
        if not self._started.is_set():
            raise RuntimeError("cannot join thread before it is started")
        if self is current_thread():
            raise RuntimeError("cannot join current thread")
        if timeout is None:
            self._done_lock.acquire()
            self._done_lock.release()
        else:
            if self._done_lock.acquire(True, max(timeout, 0)):
                self._done_lock.release()

    @property
    def name(self):
        return self._name

    @name.setter
    def name(self, name):
        self._name = str(name)

    @property
    def ident(self):
        return self._ident

    @property
    def native_id(self):
        return self._native_id

    def is_alive(self):
        return self._started.is_set() and not self._is_stopped

    @property
    def daemon(self):
        return self._daemonic

    @daemon.setter
    def daemon(self, daemonic):
        if not self._initialized:
            raise RuntimeError("Thread.__init__() not called")
        if self._started.is_set():
            raise RuntimeError("cannot set daemon status of active thread")
        self._daemonic = daemonic

    def isDaemon(self):
        return self.daemon

    def setDaemon(self, daemonic):
        self.daemon = daemonic

    def getName(self):
        return self.name

    def setName(self, name):
        self.name = name


class Timer(Thread):
    def __init__(self, interval, function, args=None, kwargs=None):
        Thread.__init__(self)
        self.interval = interval
        self.function = function
        if args is None:
            args = []
        if kwargs is None:
            kwargs = {}
        self.args = args
        self.kwargs = kwargs
        self.finished = Event()

    def cancel(self):
        self.finished.set()

    def run(self):
        self.finished.wait(self.interval)
        if not self.finished.is_set():
            self.function(*self.args, **self.kwargs)
        self.finished.set()


class _MainThread(Thread):
    def __init__(self):
        Thread.__init__(self, name="MainThread", daemon=False)
        self._ident = _native("thread.main_ident")
        self._native_id = self._ident
        self._started.set()
        self._done_lock.acquire()
        _st.active[self._ident] = self


class _DummyThread(Thread):
    def __init__(self):
        Thread.__init__(self, name=_newname("Dummy-"), daemon=True)
        self._ident = get_ident()
        self._native_id = self._ident
        self._started.set()
        _st.active[self._ident] = self


def current_thread():
    t = _st.active.get(get_ident())
    if t is not None:
        return t
    return _DummyThread()


currentThread = current_thread


def main_thread():
    return _st.main


def active_count():
    return len(_st.active) + len(_st.limbo)


activeCount = active_count


def enumerate():
    out = []
    for t in _st.active.values():
        out.append(t)
    for t in _st.limbo.values():
        out.append(t)
    return out


def _register_atexit(func, *args):
    _st.atexits.append(func)


def _shutdown():
    for f in _st.atexits:
        f()
    me = current_thread()
    while True:
        pending = []
        for t in _st.active.values():
            if t is not me and not t.daemon and t.is_alive():
                pending.append(t)
        if not pending:
            break
        for t in pending:
            t.join()


_st.main = _MainThread()
_native("thread.set_shutdown_hook", _shutdown)
