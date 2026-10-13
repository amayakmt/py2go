"""asyncio: futures, tasks and an event loop on top of the coroutines of py2go."""

import heapq
import sys
import time
import contextvars
import threading
import functools
import concurrent.futures
from collections import deque

from builtins import asyncio__exceptions_pm_CancelledError as CancelledError
from builtins import asyncio__exceptions_pm_InvalidStateError as InvalidStateError
from builtins import asyncio__exceptions_pm_BrokenBarrierError as BrokenBarrierError
from builtins import asyncio__queues_pm_QueueEmpty as QueueEmpty
from builtins import asyncio__queues_pm_QueueFull as QueueFull
from builtins import asyncio__queues_pm_QueueShutDown as QueueShutDown

from builtins import TimeoutError as TimeoutError

FIRST_COMPLETED = "FIRST_COMPLETED"
FIRST_EXCEPTION = "FIRST_EXCEPTION"
ALL_COMPLETED = "ALL_COMPLETED"

_PENDING = "PENDING"
_CANCELLED = "CANCELLED"
_FINISHED = "FINISHED"

class _State:
    def __init__(self):
        self.running_loop = None
        self.current_loop = None
        self.current_tasks = {}
        self.all_tasks = []
        self.task_counter = 0


_state = _State()


def _make_cancelled(msg):
    if msg is None:
        return CancelledError()
    return CancelledError(msg)


class Handle:
    def __init__(self, callback, args, loop, context=None):
        self._callback = callback
        self._args = args
        self._loop = loop
        self._cancelled = False
        if context is None:
            context = contextvars.copy_context()
        self._context = context

    def cancel(self):
        self._cancelled = True

    def cancelled(self):
        return self._cancelled

    def _run(self):
        try:
            self._context.run(self._callback, *self._args)
        except (SystemExit, KeyboardInterrupt):
            raise
        except BaseException as exc:
            self._loop.call_exception_handler({"message": "Exception in callback", "exception": exc, "handle": self})


class TimerHandle(Handle):
    def __init__(self, when, callback, args, loop, context=None):
        Handle.__init__(self, callback, args, loop, context)
        self._when = when

    def when(self):
        return self._when


class Future:
    def __init__(self, loop=None):
        if loop is None:
            loop = get_event_loop()
        self._loop = loop
        self._state = _PENDING
        self._result = None
        self._exception = None
        self._callbacks = []
        self._cancel_message = None
        self._cancelled_exc = None
        self._asyncio_future_blocking = False
        self._log_traceback = False

    def get_loop(self):
        return self._loop

    def _make_cancelled_error(self):
        if self._cancelled_exc is not None:
            exc = self._cancelled_exc
            self._cancelled_exc = None
            return exc
        return _make_cancelled(self._cancel_message)

    def done(self):
        return self._state != _PENDING

    def cancelled(self):
        return self._state == _CANCELLED

    def _schedule_callbacks(self):
        callbacks = self._callbacks
        self._callbacks = []
        for item in callbacks:
            self._loop.call_soon(item[0], self, context=item[1])

    def result(self):
        if self._state == _CANCELLED:
            raise self._make_cancelled_error()
        if self._state != _FINISHED:
            raise InvalidStateError("Result is not set.")
        self._log_traceback = False
        if self._exception is not None:
            raise self._exception
        return self._result

    def exception(self):
        if self._state == _CANCELLED:
            raise self._make_cancelled_error()
        if self._state != _FINISHED:
            raise InvalidStateError("Exception is not set.")
        self._log_traceback = False
        return self._exception

    def add_done_callback(self, fn, *, context=None):
        if context is None:
            context = contextvars.copy_context()
        if self._state != _PENDING:
            self._loop.call_soon(fn, self, context=context)
        else:
            self._callbacks.append((fn, context))

    def remove_done_callback(self, fn):
        kept = []
        for item in self._callbacks:
            if item[0] != fn:
                kept.append(item)
        removed = len(self._callbacks) - len(kept)
        self._callbacks = kept
        return removed

    def cancel(self, msg=None):
        if self._state != _PENDING:
            return False
        self._state = _CANCELLED
        self._cancel_message = msg
        self._schedule_callbacks()
        return True

    def set_result(self, result):
        if self._state != _PENDING:
            raise InvalidStateError("invalid state")
        self._result = result
        self._state = _FINISHED
        self._schedule_callbacks()

    def set_exception(self, exception):
        if self._state != _PENDING:
            raise InvalidStateError("invalid state")
        if _native("asyncio.isclass", exception):
            exception = exception()
        if isinstance(exception, StopIteration):
            raise TypeError("StopIteration interacts badly with generators and cannot be raised into a Future")
        self._exception = exception
        self._state = _FINISHED
        self._log_traceback = True
        self._schedule_callbacks()

    def __await__(self):
        if self._state == _PENDING:
            self._asyncio_future_blocking = True
            yield self
        if self._state == _PENDING:
            raise RuntimeError("await wasn't used with future")
        return self.result()

    def __repr__(self):
        return "<" + type(self).__name__ + " " + self._state.lower() + ">"


class Task(Future):
    def __init__(self, coro, loop=None, name=None, context=None):
        Future.__init__(self, loop)
        if not iscoroutine(coro):
            raise TypeError("a coroutine was expected, got " + repr(coro))
        self._coro = coro
        if context is None:
            context = contextvars.copy_context()
        self._context = context
        self._fut_waiter = None
        self._must_cancel = False
        self._num_cancels_requested = 0
        _state.task_counter += 1
        if name is None:
            self._name = "Task-" + str(_state.task_counter)
        else:
            self._name = str(name)
        _state.all_tasks.append(self)
        self._loop.call_soon(self._step, context=self._context)

    def get_coro(self):
        return self._coro

    def get_name(self):
        return self._name

    def set_name(self, value):
        self._name = str(value)

    def cancelling(self):
        return self._num_cancels_requested

    def uncancel(self):
        if self._num_cancels_requested > 0:
            self._num_cancels_requested -= 1
            if self._num_cancels_requested == 0:
                self._must_cancel = False
        return self._num_cancels_requested

    def cancel(self, msg=None):
        if self.done():
            return False
        self._num_cancels_requested += 1
        if self._fut_waiter is not None:
            if self._fut_waiter.cancel(msg):
                return True
        self._must_cancel = True
        self._cancel_message = msg
        return True

    def _step(self, exc=None):
        if self.done():
            raise InvalidStateError("__step(): already done")
        if self._must_cancel:
            if not isinstance(exc, CancelledError):
                exc = _make_cancelled(self._cancel_message)
            self._must_cancel = False
        self._fut_waiter = None
        _state.current_tasks[id(self._loop)] = self
        try:
            if exc is None:
                result = self._coro.send(None)
            else:
                result = self._coro.throw(exc)
        except StopIteration as e:
            if self._must_cancel:
                self._must_cancel = False
                Future.cancel(self, self._cancel_message)
            else:
                Future.set_result(self, e.value)
        except CancelledError as e:
            self._cancelled_exc = e
            Future.cancel(self, self._cancel_message)
        except (KeyboardInterrupt, SystemExit) as e:
            Future.set_exception(self, e)
            raise
        except BaseException as e:
            Future.set_exception(self, e)
        else:
            if isinstance(result, Future):
                if result._loop is not self._loop:
                    self._loop.call_soon(self._step, RuntimeError("Task " + repr(self) + " got Future " + repr(result) + " attached to a different loop"))
                elif result._asyncio_future_blocking:
                    result._asyncio_future_blocking = False
                    result.add_done_callback(self._wakeup, context=self._context)
                    self._fut_waiter = result
                    if self._must_cancel:
                        if self._fut_waiter.cancel(self._cancel_message):
                            self._must_cancel = False
                else:
                    self._loop.call_soon(self._step, RuntimeError("yield was used instead of yield from in task " + repr(self) + " with " + repr(result)))
            elif result is None:
                self._loop.call_soon(self._step, context=self._context)
            else:
                self._loop.call_soon(self._step, RuntimeError("Task got bad yield: " + repr(result)))
        finally:
            _state.current_tasks[id(self._loop)] = None

    def _wakeup(self, future):
        try:
            future.result()
        except BaseException as exc:
            self._step(exc)
        else:
            self._step()

    def __repr__(self):
        return "<Task " + self._state.lower() + " name='" + self._name + "'>"


class _GatheringFuture(Future):
    def __init__(self, children, loop):
        Future.__init__(self, loop)
        self._children = children
        self._cancel_requested = False

    def cancel(self, msg=None):
        if self.done():
            return False
        ret = False
        for child in self._children:
            if child.cancel(msg):
                ret = True
        if ret:
            self._cancel_requested = True
        return ret


class _Box:
    def __init__(self):
        self.waiter = None


class _Sleep0:
    def __await__(self):
        yield


class EventLoop:
    def __init__(self):
        self._ready = deque()
        self._scheduled = []
        self._seq = 0
        self._stopping = False
        self._running = False
        self._closed = False
        self._debug = False
        self._exception_handler = None
        self._asyncgens = []
        self._default_executor = None
        self._wake = threading.Event()

    def time(self):
        return time.monotonic()

    def is_running(self):
        return self._running

    def is_closed(self):
        return self._closed

    def get_debug(self):
        return self._debug

    def set_debug(self, enabled):
        self._debug = enabled

    def _check_closed(self):
        if self._closed:
            raise RuntimeError("Event loop is closed")

    def call_soon(self, callback, *args, context=None):
        self._check_closed()
        handle = Handle(callback, args, self, context)
        self._ready.append(handle)
        return handle

    def call_soon_threadsafe(self, callback, *args, context=None):
        handle = self.call_soon(callback, *args, context=context)
        self._wakeup_loop()
        return handle

    def _wakeup_loop(self):
        self._wake.set()

    def run_in_executor(self, executor, func, *args):
        self._check_closed()
        if executor is None:
            if self._default_executor is None:
                self._default_executor = concurrent.futures.ThreadPoolExecutor(thread_name_prefix="asyncio")
            executor = self._default_executor
        return wrap_future(executor.submit(func, *args), loop=self)

    def set_default_executor(self, executor):
        self._default_executor = executor

    def call_at(self, when, callback, *args, context=None):
        self._check_closed()
        handle = TimerHandle(when, callback, args, self, context)
        self._seq += 1
        heapq.heappush(self._scheduled, (when, self._seq, handle))
        return handle

    def call_later(self, delay, callback, *args, context=None):
        return self.call_at(self.time() + delay, callback, *args, context=context)

    def create_future(self):
        return Future(self)

    def create_task(self, coro, name=None, context=None):
        self._check_closed()
        return Task(coro, self, name, context)

    def stop(self):
        self._stopping = True

    def close(self):
        if self._running:
            raise RuntimeError("Cannot close a running event loop")
        if self._closed:
            return
        self._closed = True
        self._ready.clear()
        self._scheduled = []

    def call_exception_handler(self, context):
        if self._exception_handler is not None:
            self._exception_handler(self, context)
            return
        self.default_exception_handler(context)

    def default_exception_handler(self, context):
        message = context.get("message")
        if not message:
            message = "Unhandled exception during event loop"
        exc = context.get("exception")
        text = message
        if exc is not None:
            text = text + "\n" + type(exc).__name__ + ": " + str(exc)
        sys.stderr.write(text + "\n")

    def set_exception_handler(self, handler):
        self._exception_handler = handler

    def get_exception_handler(self):
        return self._exception_handler

    def _run_once(self):
        while self._scheduled and self._scheduled[0][2]._cancelled:
            heapq.heappop(self._scheduled)
        timeout = None
        if self._ready or self._stopping:
            timeout = 0
        elif self._scheduled:
            timeout = self._scheduled[0][0] - self.time()
            if timeout < 0:
                timeout = 0
        if timeout is None or timeout > 0:
            self._wake.wait(timeout)
        self._wake.clear()
        end_time = self.time()
        while self._scheduled and self._scheduled[0][0] <= end_time:
            item = heapq.heappop(self._scheduled)
            if not item[2]._cancelled:
                self._ready.append(item[2])
        ntodo = len(self._ready)
        for i in range(ntodo):
            handle = self._ready.popleft()
            if handle._cancelled:
                continue
            handle._run()
        handle = None

    def run_forever(self):
        self._check_closed()
        if self._running:
            raise RuntimeError("This event loop is already running")
        if _state.running_loop is not None:
            raise RuntimeError("Cannot run the event loop while another loop is running")
        self._running = True
        _state.running_loop = self
        _native("asyncio.set_agen_hook", self._agen_finalizer)
        try:
            while True:
                self._run_once()
                if self._stopping:
                    break
        finally:
            self._stopping = False
            self._running = False
            _state.running_loop = None

    def run_until_complete(self, future):
        self._check_closed()
        new_task = not isfuture(future)
        future = ensure_future(future, loop=self)
        future.add_done_callback(_run_until_complete_cb)
        try:
            self.run_forever()
        except BaseException:
            if new_task and future.done() and not future.cancelled():
                future.exception()
            raise
        finally:
            future.remove_done_callback(_run_until_complete_cb)
        if not future.done():
            raise RuntimeError("Event loop stopped before Future completed.")
        return future.result()

    def _agen_finalizer(self, agen):
        if not self._closed:
            self.create_task(agen.aclose())

    async def shutdown_asyncgens(self):
        gens = self._asyncgens
        self._asyncgens = []
        for g in gens:
            await g.aclose()

    async def shutdown_default_executor(self, timeout=None):
        if self._default_executor is not None:
            ex = self._default_executor
            self._default_executor = None
            ex.shutdown(wait=True)


def _run_until_complete_cb(fut):
    fut._loop.stop()


def isfuture(obj):
    return isinstance(obj, Future)


def iscoroutine(obj):
    return type(obj).__name__ == "coroutine"


def iscoroutinefunction(func):
    return _native("asyncio.iscoroutinefunction", func)


def get_running_loop():
    if _state.running_loop is None:
        raise RuntimeError("no running event loop")
    return _state.running_loop


def _get_running_loop():
    return _state.running_loop


def get_event_loop():
    if _state.running_loop is not None:
        return _state.running_loop
    if _state.current_loop is None:
        raise RuntimeError("There is no current event loop in thread 'MainThread'.")
    return _state.current_loop


def new_event_loop():
    return EventLoop()


def set_event_loop(loop):
    _state.current_loop = loop


def current_task(loop=None):
    if loop is None:
        loop = get_running_loop()
    return _state.current_tasks.get(id(loop))


def all_tasks(loop=None):
    if loop is None:
        loop = get_running_loop()
    out = set()
    for t in _state.all_tasks:
        if t._loop is loop and not t.done():
            out.add(t)
    return out


def ensure_future(obj, *, loop=None):
    if isfuture(obj):
        return obj
    if iscoroutine(obj):
        if loop is None:
            loop = get_event_loop()
        return loop.create_task(obj)
    if hasattr(obj, "__await__"):
        return ensure_future(_wrap_awaitable(obj), loop=loop)
    raise TypeError("An asyncio.Future, a coroutine or an awaitable is required")


async def _wrap_awaitable(awaitable):
    return await awaitable


def create_task(coro, *, name=None, context=None):
    loop = get_running_loop()
    return loop.create_task(coro, name, context)


def _cancel_all_tasks(loop):
    to_cancel = []
    for t in _state.all_tasks:
        if t._loop is loop and not t.done():
            to_cancel.append(t)
    if not to_cancel:
        return
    for task in to_cancel:
        task.cancel()
    loop.run_until_complete(gather(*to_cancel, return_exceptions=True))
    for task in to_cancel:
        if task.cancelled():
            continue
        if task.exception() is not None:
            loop.call_exception_handler({"message": "unhandled exception during asyncio.run() shutdown", "exception": task.exception(), "task": task})


def run(main, *, debug=None, loop_factory=None):
    if _state.running_loop is not None:
        raise RuntimeError("asyncio.run() cannot be called from a running event loop")
    if not iscoroutine(main):
        raise ValueError("a coroutine was expected, got " + repr(main))
    _state.task_counter = 0
    loop = new_event_loop()
    set_event_loop(loop)
    try:
        return loop.run_until_complete(main)
    finally:
        try:
            _cancel_all_tasks(loop)
            loop.run_until_complete(loop.shutdown_asyncgens())
            loop.run_until_complete(loop.shutdown_default_executor())
        finally:
            set_event_loop(None)
            loop.close()


def _set_result_unless_cancelled(fut, result):
    if fut.cancelled():
        return
    fut.set_result(result)


async def sleep(delay, result=None):
    if delay <= 0:
        await _Sleep0()
        return result
    loop = get_running_loop()
    future = loop.create_future()
    h = loop.call_later(delay, _set_result_unless_cancelled, future, result)
    try:
        return await future
    finally:
        h.cancel()


def gather(*coros_or_futures, return_exceptions=False):
    if not coros_or_futures:
        loop = get_event_loop()
        outer = loop.create_future()
        outer.set_result([])
        return outer
    arg_to_fut = {}
    children = []
    nfuts = 0
    nfinished = 0
    loop = None
    outer = None
    results = [None] * len(coros_or_futures)
    done_count = [0]
    for arg in coros_or_futures:
        key = id(arg)
        if key not in arg_to_fut:
            fut = ensure_future(arg, loop=loop)
            if loop is None:
                loop = fut._loop
            arg_to_fut[key] = fut
            if fut is not arg:
                fut._log_traceback = False
        else:
            fut = arg_to_fut[key]
        children.append(fut)
    nfuts = len(arg_to_fut)
    outer = _GatheringFuture(children, loop)

    def _done_callback(fut):
        done_count[0] += 1
        if outer is None or outer.done():
            if not fut.cancelled():
                fut.exception()
            return
        if not return_exceptions:
            if fut.cancelled():
                exc = _make_cancelled(fut._cancel_message)
                outer.set_exception(exc)
                return
            else:
                exc = fut.exception()
                if exc is not None:
                    outer.set_exception(exc)
                    return
        if done_count[0] == nfuts:
            res = []
            for child in children:
                if child.cancelled():
                    res.append(_make_cancelled(child._cancel_message))
                else:
                    e = child.exception()
                    if e is not None:
                        res.append(e)
                    else:
                        res.append(child.result())
            if outer._cancel_requested:
                outer.set_exception(CancelledError())
            else:
                outer.set_result(res)

    for fut in arg_to_fut.values():
        fut.add_done_callback(_done_callback)
    return outer


def shield(arg):
    inner = ensure_future(arg)
    if inner.done():
        return inner
    loop = inner._loop
    outer = loop.create_future()

    def _inner_done_callback(inner):
        if outer.cancelled():
            if not inner.cancelled():
                inner.exception()
            return
        if inner.cancelled():
            outer.cancel()
        else:
            exc = inner.exception()
            if exc is not None:
                outer.set_exception(exc)
            else:
                outer.set_result(inner.result())

    def _outer_done_callback(outer):
        if not inner.done():
            inner.remove_done_callback(_inner_done_callback)

    inner.add_done_callback(_inner_done_callback)
    outer.add_done_callback(_outer_done_callback)
    return outer


async def _cancel_and_wait(fut):
    loop = get_running_loop()
    waiter = loop.create_future()
    cb = lambda f: _release_waiter(waiter)
    fut.add_done_callback(cb)
    try:
        fut.cancel()
        await waiter
    finally:
        fut.remove_done_callback(cb)


def _release_waiter(waiter, *args):
    if not waiter.done():
        waiter.set_result(None)


async def wait(fs, *, timeout=None, return_when=ALL_COMPLETED):
    if isfuture(fs) or iscoroutine(fs):
        raise TypeError("expect a list of futures, not " + type(fs).__name__)
    if not fs:
        raise ValueError("Set of Tasks/Futures is empty.")
    if return_when != FIRST_COMPLETED and return_when != FIRST_EXCEPTION and return_when != ALL_COMPLETED:
        raise ValueError("Invalid return_when value: " + str(return_when))
    loop = get_running_loop()
    todo = set()
    for f in fs:
        todo.add(ensure_future(f, loop=loop))
    return await _wait(todo, timeout, return_when, loop)


async def _wait(fs, timeout, return_when, loop):
    waiter = loop.create_future()
    timeout_handle = None
    if timeout is not None:
        timeout_handle = loop.call_later(timeout, _release_waiter, waiter)
    counter = [len(fs)]

    def _on_completion(f):
        counter[0] -= 1
        if counter[0] <= 0 or return_when == FIRST_COMPLETED or (return_when == FIRST_EXCEPTION and (not f.cancelled() and f.exception() is not None)):
            if timeout_handle is not None:
                timeout_handle.cancel()
            if not waiter.done():
                waiter.set_result(None)

    for f in fs:
        f.add_done_callback(_on_completion)
    try:
        await waiter
    finally:
        if timeout_handle is not None:
            timeout_handle.cancel()
        for f in fs:
            f.remove_done_callback(_on_completion)
    done = set()
    pending = set()
    for f in fs:
        if f.done():
            done.add(f)
        else:
            pending.add(f)
    return done, pending


class Timeout:
    def __init__(self, when):
        self._state = "created"
        self._timeout_handler = None
        self._task = None
        self._when = when
        self._cancelling = 0

    def when(self):
        return self._when

    def reschedule(self, when):
        self._when = when
        if self._timeout_handler is not None:
            self._timeout_handler.cancel()
            self._timeout_handler = None
        if when is None:
            return
        loop = get_running_loop()
        if when <= loop.time():
            self._timeout_handler = loop.call_soon(self._on_timeout)
        else:
            self._timeout_handler = loop.call_at(when, self._on_timeout)

    def expired(self):
        return self._state == "expired" or self._state == "expiring"

    async def __aenter__(self):
        if self._state != "created":
            raise RuntimeError("Timeout has already been entered")
        task = current_task()
        if task is None:
            raise RuntimeError("Timeout should be used inside a task")
        self._state = "entered"
        self._task = task
        self._cancelling = task.cancelling()
        self.reschedule(self._when)
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self._timeout_handler is not None:
            self._timeout_handler.cancel()
            self._timeout_handler = None
        if self._state == "expiring":
            self._state = "expired"
            if self._task.uncancel() <= self._cancelling and exc_type is not None:
                if issubclass(exc_type, CancelledError):
                    raise TimeoutError from exc_val
        elif self._state == "entered":
            self._state = "exited"
        return None

    def _on_timeout(self):
        self._task.cancel()
        self._state = "expiring"
        self._timeout_handler = None


def timeout(delay):
    loop = get_running_loop()
    if delay is None:
        return Timeout(None)
    return Timeout(loop.time() + delay)


def timeout_at(when):
    return Timeout(when)


async def wait_for(fut, timeout):
    if timeout is not None and timeout <= 0:
        fut = ensure_future(fut)
        if fut.done():
            return fut.result()
        await _cancel_and_wait(fut)
        try:
            return fut.result()
        except CancelledError as exc:
            raise TimeoutError from exc
    async with Timeout(None if timeout is None else get_running_loop().time() + timeout):
        return await fut


def as_completed(fs, *, timeout=None):
    todo = set()
    for f in fs:
        todo.add(ensure_future(f))
    loop = get_event_loop()
    done = deque()
    state = _Box()

    def _on_completion(f):
        todo.discard(f)
        done.append(f)
        w = state.waiter
        if w is not None and not w.done():
            w.set_result(None)

    for f in todo:
        f.add_done_callback(_on_completion)
    total = len(todo)

    async def _wait_for_one():
        while not done:
            if not todo:
                return None
            state.waiter = loop.create_future()
            await state.waiter
        f = done.popleft()
        return f.result()

    return [_wait_for_one() for _ in range(total)]


class TaskGroup:
    def __init__(self):
        self._entered = False
        self._exiting = False
        self._aborting = False
        self._parent_task = None
        self._parent_cancel_requested = False
        self._tasks = set()
        self._errors = []
        self._base_error = None
        self._on_completed_fut = None
        self._loop = get_event_loop()

    async def __aenter__(self):
        if self._entered:
            raise RuntimeError("TaskGroup " + repr(self) + " has already been entered")
        self._entered = True
        self._parent_task = current_task()
        if self._parent_task is None:
            raise RuntimeError("TaskGroup " + repr(self) + " cannot determine the parent task")
        return self

    async def __aexit__(self, et, exc, tb):
        self._exiting = True
        if exc is not None and self._is_base_error(exc) and self._base_error is None:
            self._base_error = exc
        if et is not None and issubclass(et, CancelledError):
            if self._parent_cancel_requested:
                self._parent_task.uncancel()
        if et is not None:
            if not self._aborting:
                self._abort()
        while self._tasks:
            if self._on_completed_fut is None:
                self._on_completed_fut = self._loop.create_future()
            try:
                await self._on_completed_fut
            except CancelledError as ex:
                if not self._aborting:
                    exc = ex
                    self._abort()
            self._on_completed_fut = None
        if self._base_error is not None:
            raise self._base_error
        if et is not None and issubclass(et, CancelledError) and not self._errors:
            raise exc
        if et is not None and not issubclass(et, CancelledError):
            self._errors.append(exc)
        if self._errors:
            errors = self._errors
            self._errors = []
            raise BaseExceptionGroup("unhandled errors in a TaskGroup", errors)

    def create_task(self, coro, *, name=None, context=None):
        if not self._entered:
            raise RuntimeError("TaskGroup " + repr(self) + " has not been entered")
        if self._exiting and not self._tasks:
            raise RuntimeError("TaskGroup " + repr(self) + " is finished")
        if self._aborting:
            raise RuntimeError("TaskGroup " + repr(self) + " is shutting down")
        task = self._loop.create_task(coro, name)
        if task.done():
            self._on_task_done(task)
        else:
            self._tasks.add(task)
            task.add_done_callback(self._on_task_done)
        return task

    def _is_base_error(self, exc):
        return isinstance(exc, (SystemExit, KeyboardInterrupt))

    def _abort(self):
        self._aborting = True
        for t in self._tasks:
            if not t.done():
                t.cancel()

    def _on_task_done(self, task):
        self._tasks.discard(task)
        if self._on_completed_fut is not None and not self._tasks:
            if not self._on_completed_fut.done():
                self._on_completed_fut.set_result(True)
        if task.cancelled():
            return
        exc = task.exception()
        if exc is None:
            return
        self._errors.append(exc)
        if self._is_base_error(exc) and self._base_error is None:
            self._base_error = exc
        if self._parent_task.done():
            self._loop.call_exception_handler({"message": "Task " + repr(task) + " has errored out but its parent task " + repr(self._parent_task) + " is already completed", "exception": exc, "task": task})
            return
        if not self._aborting and not self._parent_cancel_requested:
            self._abort()
            self._parent_cancel_requested = True
            self._parent_task.cancel()


# ---- synchronization primitives ----


class Lock:
    def __init__(self):
        self._waiters = None
        self._locked = False

    def locked(self):
        return self._locked

    async def acquire(self):
        if not self._locked and (self._waiters is None or all_cancelled(self._waiters)):
            self._locked = True
            return True
        if self._waiters is None:
            self._waiters = deque()
        fut = get_running_loop().create_future()
        self._waiters.append(fut)
        try:
            try:
                await fut
            finally:
                self._waiters.remove(fut)
        except CancelledError:
            if not self._locked:
                self._wake_up_first()
            raise
        self._locked = True
        return True

    def release(self):
        if self._locked:
            self._locked = False
            self._wake_up_first()
        else:
            raise RuntimeError("Lock is not acquired.")

    def _wake_up_first(self):
        if not self._waiters:
            return
        for fut in self._waiters:
            if not fut.done():
                fut.set_result(True)
                return

    async def __aenter__(self):
        await self.acquire()
        return None

    async def __aexit__(self, exc_type, exc, tb):
        self.release()


def all_cancelled(waiters):
    for w in waiters:
        if not w.cancelled():
            return False
    return True


class Event:
    def __init__(self):
        self._waiters = deque()
        self._value = False

    def is_set(self):
        return self._value

    def set(self):
        if not self._value:
            self._value = True
            for fut in self._waiters:
                if not fut.done():
                    fut.set_result(True)

    def clear(self):
        self._value = False

    async def wait(self):
        if self._value:
            return True
        fut = get_running_loop().create_future()
        self._waiters.append(fut)
        try:
            await fut
            return True
        finally:
            self._waiters.remove(fut)


class Condition:
    def __init__(self, lock=None):
        if lock is None:
            lock = Lock()
        self._lock = lock
        self._waiters = deque()

    def locked(self):
        return self._lock.locked()

    async def acquire(self):
        return await self._lock.acquire()

    def release(self):
        self._lock.release()

    async def __aenter__(self):
        await self.acquire()
        return None

    async def __aexit__(self, exc_type, exc, tb):
        self.release()

    async def wait(self):
        if not self._lock.locked():
            raise RuntimeError("cannot wait on un-acquired lock")
        self.release()
        try:
            fut = get_running_loop().create_future()
            self._waiters.append(fut)
            try:
                await fut
                return True
            finally:
                self._waiters.remove(fut)
        finally:
            cancelled = False
            while True:
                try:
                    await self.acquire()
                    break
                except CancelledError:
                    cancelled = True
            if cancelled:
                raise CancelledError()

    async def wait_for(self, predicate):
        result = predicate()
        while not result:
            await self.wait()
            result = predicate()
        return result

    def notify(self, n=1):
        if not self._lock.locked():
            raise RuntimeError("cannot notify on un-acquired lock")
        idx = 0
        for fut in self._waiters:
            if idx >= n:
                break
            if not fut.done():
                idx += 1
                fut.set_result(False)

    def notify_all(self):
        self.notify(len(self._waiters))


class Semaphore:
    def __init__(self, value=1):
        if value < 0:
            raise ValueError("Semaphore initial value must be >= 0")
        self._waiters = None
        self._value = value

    def locked(self):
        return self._value == 0 or (self._waiters is not None and not all_cancelled(self._waiters))

    async def acquire(self):
        if not self.locked():
            self._value -= 1
            return True
        if self._waiters is None:
            self._waiters = deque()
        fut = get_running_loop().create_future()
        self._waiters.append(fut)
        try:
            try:
                await fut
            finally:
                self._waiters.remove(fut)
        except CancelledError:
            if not fut.cancelled():
                self._value += 1
                self._wake_up_next()
            raise
        if self._value > 0:
            self._wake_up_next()
        return True

    def release(self):
        self._value += 1
        self._wake_up_next()

    def _wake_up_next(self):
        if not self._waiters:
            return False
        while self._value > 0:
            for fut in self._waiters:
                if not fut.done():
                    self._value -= 1
                    fut.set_result(True)
                    return True
            return False
        return False

    async def __aenter__(self):
        await self.acquire()
        return None

    async def __aexit__(self, exc_type, exc, tb):
        self.release()


class BoundedSemaphore(Semaphore):
    def __init__(self, value=1):
        self._bound_value = value
        Semaphore.__init__(self, value)

    def release(self):
        if self._value >= self._bound_value:
            raise ValueError("BoundedSemaphore released too many times")
        Semaphore.release(self)


class Barrier:
    def __init__(self, parties):
        if parties < 1:
            raise ValueError("parties must be >= 1")
        self._cond = Condition()
        self._parties = parties
        self._state = 0
        self._count = 0

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

    async def __aenter__(self):
        await self.wait()
        return self

    async def __aexit__(self, exc_type, exc, tb):
        pass

    async def wait(self):
        async with self._cond:
            await self._block()
            try:
                index = self._count
                self._count += 1
                if index + 1 == self._parties:
                    await self._release()
                else:
                    await self._wait()
                return index
            finally:
                self._count -= 1
                self._exit()

    async def _block(self):
        await self._cond.wait_for(lambda: self._state not in (1, -1))
        if self._state == -2:
            raise BrokenBarrierError

    async def _release(self):
        self._state = 1
        self._cond.notify_all()

    async def _wait(self):
        await self._cond.wait_for(lambda: self._state != 0)
        if self._state < 0:
            raise BrokenBarrierError

    def _exit(self):
        if self._count == 0:
            if self._state in (1, -1):
                self._state = 0
                self._cond.notify_all()

    async def reset(self):
        async with self._cond:
            if self._count > 0:
                if self._state != -1:
                    self._state = -1
            else:
                self._state = 0
            self._cond.notify_all()

    async def abort(self):
        async with self._cond:
            self._state = -2
            self._cond.notify_all()


class Queue:
    def __init__(self, maxsize=0):
        self._maxsize = maxsize
        self._getters = deque()
        self._putters = deque()
        self._unfinished_tasks = 0
        self._finished = Event()
        self._finished.set()
        self._is_shutdown = False
        self._init(maxsize)

    def _init(self, maxsize):
        self._queue = deque()

    def _get(self):
        return self._queue.popleft()

    def _put(self, item):
        self._queue.append(item)

    def _wakeup_next(self, waiters):
        while waiters:
            waiter = waiters.popleft()
            if not waiter.done():
                waiter.set_result(None)
                break

    def __repr__(self):
        return "<" + type(self).__name__ + " " + self._format() + ">"

    def _format(self):
        result = "maxsize=" + repr(self._maxsize)
        if self._qsize():
            result += " _queue=" + repr(self._items())
        if self._getters:
            result += " _getters[" + str(len(self._getters)) + "]"
        if self._putters:
            result += " _putters[" + str(len(self._putters)) + "]"
        if self._unfinished_tasks:
            result += " tasks=" + str(self._unfinished_tasks)
        return result

    def _qsize(self):
        return len(self._queue)

    def _items(self):
        return list(self._queue)

    def qsize(self):
        return self._qsize()

    @property
    def maxsize(self):
        return self._maxsize

    def empty(self):
        return self._qsize() == 0

    def full(self):
        if self._maxsize <= 0:
            return False
        return self.qsize() >= self._maxsize

    async def put(self, item):
        while self.full():
            if self._is_shutdown:
                raise QueueShutDown
            putter = get_running_loop().create_future()
            self._putters.append(putter)
            try:
                await putter
            except BaseException:
                if putter in self._putters:
                    self._putters.remove(putter)
                if not self.full() and not putter.cancelled():
                    self._wakeup_next(self._putters)
                raise
        return self.put_nowait(item)

    def put_nowait(self, item):
        if self._is_shutdown:
            raise QueueShutDown
        if self.full():
            raise QueueFull
        self._put(item)
        self._unfinished_tasks += 1
        self._finished.clear()
        self._wakeup_next(self._getters)

    async def get(self):
        while self.empty():
            if self._is_shutdown:
                raise QueueShutDown
            getter = get_running_loop().create_future()
            self._getters.append(getter)
            try:
                await getter
            except BaseException:
                if getter in self._getters:
                    self._getters.remove(getter)
                if not self.empty() and not getter.cancelled():
                    self._wakeup_next(self._getters)
                raise
        return self.get_nowait()

    def get_nowait(self):
        if self.empty():
            if self._is_shutdown:
                raise QueueShutDown
            raise QueueEmpty
        item = self._get()
        self._wakeup_next(self._putters)
        return item

    def task_done(self):
        if self._unfinished_tasks <= 0:
            raise ValueError("task_done() called too many times")
        self._unfinished_tasks -= 1
        if self._unfinished_tasks == 0:
            self._finished.set()

    async def join(self):
        if self._unfinished_tasks > 0:
            await self._finished.wait()

    def shutdown(self, immediate=False):
        self._is_shutdown = True
        if immediate:
            while not self.empty():
                self._get()
                if self._unfinished_tasks > 0:
                    self._unfinished_tasks -= 1
            if self._unfinished_tasks == 0:
                self._finished.set()
        for getters in (self._getters, self._putters):
            while getters:
                waiter = getters.popleft()
                if not waiter.done():
                    waiter.set_result(None)


class PriorityQueue(Queue):
    def _init(self, maxsize):
        self._heap = []

    def _qsize(self):
        return len(self._heap)

    def _items(self):
        return list(self._heap)

    def _put(self, item):
        heapq.heappush(self._heap, item)

    def _get(self):
        return heapq.heappop(self._heap)


class LifoQueue(Queue):
    def _init(self, maxsize):
        self._stack = []

    def _qsize(self):
        return len(self._stack)

    def _items(self):
        return list(self._stack)

    def _put(self, item):
        self._stack.append(item)

    def _get(self):
        return self._stack.pop()


def wrap_future(future, *, loop=None):
    if isfuture(future):
        return future
    if loop is None:
        loop = get_event_loop()
    new_future = loop.create_future()

    def _copy(source):
        if new_future.cancelled():
            return
        if source.cancelled():
            new_future.cancel()
            return
        exc = source.exception()
        if exc is not None:
            new_future.set_exception(exc)
        else:
            new_future.set_result(source.result())

    def _call_check_cancel(destination):
        if destination.cancelled():
            future.cancel()

    def _call_set_state(source):
        if new_future.cancelled() and loop.is_closed():
            return
        loop.call_soon_threadsafe(_copy, source)

    new_future.add_done_callback(_call_check_cancel)
    future.add_done_callback(_call_set_state)
    return new_future


async def to_thread(func, /, *args, **kwargs):
    loop = get_running_loop()
    ctx = contextvars.copy_context()
    func_call = functools.partial(ctx.run, func, *args, **kwargs)
    return await loop.run_in_executor(None, func_call)


def run_coroutine_threadsafe(coro, loop):
    if not iscoroutine(coro):
        raise TypeError("A coroutine object is required")
    future = concurrent.futures.Future()

    def callback():
        try:
            fut = ensure_future(coro, loop=loop)

            def _done(f):
                if future.cancelled():
                    return
                if f.cancelled():
                    future.cancel()
                    return
                exc = f.exception()
                if exc is not None:
                    future.set_exception(exc)
                else:
                    future.set_result(f.result())

            fut.add_done_callback(_done)
        except (SystemExit, KeyboardInterrupt):
            raise
        except BaseException as exc:
            if future.set_running_or_notify_cancel():
                future.set_exception(exc)
            raise

    loop.call_soon_threadsafe(callback)
    return future
