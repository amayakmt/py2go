"""concurrent.futures: futures and a thread pool executor."""

import threading
import itertools
import time
from collections import deque
import queue

from builtins import concurrent__futures___base_pm_Error as Error
from builtins import concurrent__futures___base_pm_CancelledError as CancelledError
from builtins import concurrent__futures___base_pm_InvalidStateError as InvalidStateError
from builtins import concurrent__futures___base_pm_BrokenExecutor as BrokenExecutor
from builtins import concurrent__futures__thread_pm_BrokenThreadPool as BrokenThreadPool
from builtins import TimeoutError as TimeoutError

FIRST_COMPLETED = "FIRST_COMPLETED"
FIRST_EXCEPTION = "FIRST_EXCEPTION"
ALL_COMPLETED = "ALL_COMPLETED"

_PENDING = "PENDING"
_RUNNING = "RUNNING"
_CANCELLED = "CANCELLED"
_CANCELLED_AND_NOTIFIED = "CANCELLED_AND_NOTIFIED"
_FINISHED = "FINISHED"


class Future:
    def __init__(self):
        self._condition = threading.Condition()
        self._state = _PENDING
        self._result = None
        self._exception = None
        self._waiters = []
        self._done_callbacks = []

    def _invoke_callbacks(self):
        for callback in self._done_callbacks:
            try:
                callback(self)
            except Exception as e:
                pass

    def __repr__(self):
        with self._condition:
            if self._state == _FINISHED:
                if self._exception is not None:
                    return "<Future at " + hex(id(self)) + " state=finished raised " + type(self._exception).__name__ + ">"
                return "<Future at " + hex(id(self)) + " state=finished returned " + type(self._result).__name__ + ">"
            return "<Future at " + hex(id(self)) + " state=" + self._state.lower() + ">"

    def cancel(self):
        with self._condition:
            if self._state == _RUNNING or self._state == _FINISHED:
                return False
            if self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED:
                return True
            self._state = _CANCELLED
            self._condition.notify_all()
        self._invoke_callbacks()
        return True

    def cancelled(self):
        with self._condition:
            return self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED

    def running(self):
        with self._condition:
            return self._state == _RUNNING

    def done(self):
        with self._condition:
            return self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED or self._state == _FINISHED

    def __get_result(self):
        if self._exception is not None:
            raise self._exception
        return self._result

    def add_done_callback(self, fn):
        with self._condition:
            if self._state != _CANCELLED and self._state != _CANCELLED_AND_NOTIFIED and self._state != _FINISHED:
                self._done_callbacks.append(fn)
                return
        try:
            fn(self)
        except Exception as e:
            pass

    def result(self, timeout=None):
        with self._condition:
            if self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED:
                raise CancelledError()
            elif self._state == _FINISHED:
                return self.__get_result()
            self._condition.wait(timeout)
            if self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED:
                raise CancelledError()
            elif self._state == _FINISHED:
                return self.__get_result()
            else:
                raise TimeoutError()

    def exception(self, timeout=None):
        with self._condition:
            if self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED:
                raise CancelledError()
            elif self._state == _FINISHED:
                return self._exception
            self._condition.wait(timeout)
            if self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED:
                raise CancelledError()
            elif self._state == _FINISHED:
                return self._exception
            else:
                raise TimeoutError()

    def set_running_or_notify_cancel(self):
        with self._condition:
            if self._state == _CANCELLED:
                self._state = _CANCELLED_AND_NOTIFIED
                for waiter in self._waiters:
                    waiter.add_cancelled(self)
                return False
            elif self._state == _PENDING:
                self._state = _RUNNING
                return True
            else:
                raise RuntimeError("Future in unexpected state")

    def set_result(self, result):
        with self._condition:
            if self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED or self._state == _FINISHED:
                raise InvalidStateError(self._state + ": " + repr(self))
            self._result = result
            self._state = _FINISHED
            for waiter in self._waiters:
                waiter.add_result(self)
            self._condition.notify_all()
        self._invoke_callbacks()

    def set_exception(self, exception):
        with self._condition:
            if self._state == _CANCELLED or self._state == _CANCELLED_AND_NOTIFIED or self._state == _FINISHED:
                raise InvalidStateError(self._state + ": " + repr(self))
            self._exception = exception
            self._state = _FINISHED
            for waiter in self._waiters:
                waiter.add_exception(self)
            self._condition.notify_all()
        self._invoke_callbacks()


class _Waiter:
    def __init__(self):
        self.event = threading.Event()
        self.finished_futures = []

    def add_result(self, future):
        self.finished_futures.append(future)

    def add_exception(self, future):
        self.finished_futures.append(future)

    def add_cancelled(self, future):
        self.finished_futures.append(future)


class _AsCompletedWaiter(_Waiter):
    def __init__(self):
        _Waiter.__init__(self)
        self.lock = threading.Lock()

    def add_result(self, future):
        with self.lock:
            _Waiter.add_result(self, future)
            self.event.set()

    def add_exception(self, future):
        with self.lock:
            _Waiter.add_exception(self, future)
            self.event.set()

    def add_cancelled(self, future):
        with self.lock:
            _Waiter.add_cancelled(self, future)
            self.event.set()


class _FirstCompletedWaiter(_Waiter):
    def add_result(self, future):
        _Waiter.add_result(self, future)
        self.event.set()

    def add_exception(self, future):
        _Waiter.add_exception(self, future)
        self.event.set()

    def add_cancelled(self, future):
        _Waiter.add_cancelled(self, future)
        self.event.set()


class _AllCompletedWaiter(_Waiter):
    def __init__(self, num_pending_calls, stop_on_exception):
        self.num_pending_calls = num_pending_calls
        self.stop_on_exception = stop_on_exception
        self.lock = threading.Lock()
        _Waiter.__init__(self)

    def _decrement_pending_calls(self):
        with self.lock:
            self.num_pending_calls -= 1
            if not self.num_pending_calls:
                self.event.set()

    def add_result(self, future):
        _Waiter.add_result(self, future)
        self._decrement_pending_calls()

    def add_exception(self, future):
        _Waiter.add_exception(self, future)
        if self.stop_on_exception:
            self.event.set()
        else:
            self._decrement_pending_calls()

    def add_cancelled(self, future):
        _Waiter.add_cancelled(self, future)
        self._decrement_pending_calls()


def _create_and_install_waiters(fs, return_when):
    if return_when == "_AS_COMPLETED":
        waiter = _AsCompletedWaiter()
    elif return_when == FIRST_COMPLETED:
        waiter = _FirstCompletedWaiter()
    else:
        pending_count = 0
        for f in fs:
            if f._state != _CANCELLED_AND_NOTIFIED and f._state != _FINISHED:
                pending_count += 1
        if return_when == FIRST_EXCEPTION:
            waiter = _AllCompletedWaiter(pending_count, True)
        elif return_when == ALL_COMPLETED:
            waiter = _AllCompletedWaiter(pending_count, False)
        else:
            raise ValueError("Invalid return condition: " + repr(return_when))
    for f in fs:
        f._waiters.append(waiter)
    return waiter


class DoneAndNotDoneFutures:
    def __init__(self, done, not_done):
        self.done = done
        self.not_done = not_done

    def __iter__(self):
        return iter([self.done, self.not_done])

    def __getitem__(self, i):
        if i == 0:
            return self.done
        return self.not_done

    def __repr__(self):
        return "DoneAndNotDoneFutures(done=" + repr(self.done) + ", not_done=" + repr(self.not_done) + ")"


def wait(fs, timeout=None, return_when=ALL_COMPLETED):
    fs = set(fs)
    done = set()
    for f in fs:
        if f._state == _CANCELLED_AND_NOTIFIED or f._state == _FINISHED:
            done.add(f)
    not_done = fs - done
    if return_when == FIRST_COMPLETED and done:
        return DoneAndNotDoneFutures(done, not_done)
    if return_when == FIRST_EXCEPTION and done:
        for f in done:
            if not f.cancelled() and f.exception() is not None:
                return DoneAndNotDoneFutures(done, not_done)
    if len(done) == len(fs):
        return DoneAndNotDoneFutures(done, not_done)
    waiter = _create_and_install_waiters(fs, return_when)
    waiter.event.wait(timeout)
    for f in fs:
        f._waiters.remove(waiter)
    done.update(waiter.finished_futures)
    return DoneAndNotDoneFutures(done, fs - done)


def as_completed(fs, timeout=None):
    if timeout is not None:
        end_time = timeout + time.monotonic()
    fs = set(fs)
    total_futures = len(fs)
    finished = set()
    for f in fs:
        if f._state == _CANCELLED_AND_NOTIFIED or f._state == _FINISHED:
            finished.add(f)
    pending = fs - finished
    waiter = _create_and_install_waiters(fs, "_AS_COMPLETED")
    finished = list(finished)
    try:
        for future in finished:
            yield future
        while pending:
            if timeout is None:
                wait_timeout = None
            else:
                wait_timeout = end_time - time.monotonic()
                if wait_timeout < 0:
                    raise TimeoutError(str(len(pending)) + " (of " + str(total_futures) + ") futures unfinished")
            waiter.event.wait(wait_timeout)
            with waiter.lock:
                finished = waiter.finished_futures
                waiter.finished_futures = []
                waiter.event.clear()
            finished.reverse()
            for future in finished:
                yield future
                pending.remove(future)
    finally:
        for f in fs:
            f._waiters.remove(waiter)


class Executor:
    def submit(self, fn, /, *args, **kwargs):
        raise NotImplementedError()

    def map(self, fn, *iterables, timeout=None, chunksize=1):
        if timeout is not None:
            end_time = timeout + time.monotonic()
        fs = []
        for args in zip(*iterables):
            fs.append(self.submit(fn, *args))

        def result_iterator():
            try:
                fs.reverse()
                while fs:
                    if timeout is None:
                        yield fs.pop().result()
                    else:
                        yield fs.pop().result(end_time - time.monotonic())
            finally:
                for future in fs:
                    future.cancel()

        return result_iterator()

    def shutdown(self, wait=True, *, cancel_futures=False):
        pass

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.shutdown(wait=True)
        return False


class _WorkItem:
    def __init__(self, future, fn, args, kwargs):
        self.future = future
        self.fn = fn
        self.args = args
        self.kwargs = kwargs

    def run(self):
        if not self.future.set_running_or_notify_cancel():
            return
        try:
            result = self.fn(*self.args, **self.kwargs)
        except BaseException as exc:
            self.future.set_exception(exc)
        else:
            self.future.set_result(result)


class _ExState:
    def __init__(self):
        self.shutdown = False
        self.queues = []
        self.threads = []
        self.counter = 0


_ex = _ExState()


def _python_exit():
    _ex.shutdown = True
    for q in _ex.queues:
        q.put(None)
    for t in _ex.threads:
        t.join()


threading._register_atexit(_python_exit)


def _worker(executor, work_queue, initializer, initargs):
    if initializer is not None:
        try:
            initializer(*initargs)
        except BaseException:
            executor._broken = "A thread initializer failed, the thread pool is not usable anymore"
            while True:
                try:
                    w = work_queue.get_nowait()
                except queue.Empty:
                    break
                if w is not None:
                    w.future.set_exception(BrokenThreadPool(executor._broken))
            executor._shutdown = True
            return
    while True:
        work_item = work_queue.get(block=True)
        if work_item is not None:
            work_item.run()
            work_item = None
            executor._idle_semaphore.release()
            continue
        if _ex.shutdown or executor._shutdown:
            work_queue.put(None)
            return


class ThreadPoolExecutor(Executor):
    def __init__(self, max_workers=None, thread_name_prefix="", initializer=None, initargs=()):
        if max_workers is None:
            max_workers = 12
        if max_workers <= 0:
            raise ValueError("max_workers must be greater than 0")
        if initializer is not None and not callable(initializer):
            raise TypeError("initializer must be a callable")
        self._max_workers = max_workers
        self._work_queue = queue.SimpleQueue()
        self._idle_semaphore = threading.Semaphore(0)
        self._threads = []
        self._broken = ""
        self._shutdown = False
        self._shutdown_lock = threading.Lock()
        self._thread_name_prefix = thread_name_prefix
        if not thread_name_prefix:
            _ex.counter += 1
            self._thread_name_prefix = "ThreadPoolExecutor-" + str(_ex.counter - 1)
        self._initializer = initializer
        self._initargs = initargs
        _ex.queues.append(self._work_queue)

    def submit(self, fn, /, *args, **kwargs):
        with self._shutdown_lock:
            if self._broken:
                raise BrokenThreadPool(self._broken)
            if self._shutdown:
                raise RuntimeError("cannot schedule new futures after shutdown")
            if _ex.shutdown:
                raise RuntimeError("cannot schedule new futures after interpreter shutdown")
            f = Future()
            w = _WorkItem(f, fn, args, kwargs)
            self._work_queue.put(w)
            self._adjust_thread_count()
            return f

    def _adjust_thread_count(self):
        if self._idle_semaphore.acquire(False):
            return
        num_threads = len(self._threads)
        if num_threads < self._max_workers:
            thread_name = self._thread_name_prefix + "_" + str(num_threads)
            t = threading.Thread(name=thread_name, target=_worker, args=(self, self._work_queue, self._initializer, self._initargs))
            t.start()
            self._threads.append(t)
            _ex.threads.append(t)

    def shutdown(self, wait=True, *, cancel_futures=False):
        with self._shutdown_lock:
            self._shutdown = True
            if cancel_futures:
                while True:
                    try:
                        work_item = self._work_queue.get_nowait()
                    except queue.Empty:
                        break
                    if work_item is not None:
                        work_item.future.cancel()
            self._work_queue.put(None)
        if wait:
            for t in self._threads:
                t.join()
