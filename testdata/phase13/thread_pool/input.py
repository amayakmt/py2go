import time
import threading
from concurrent.futures import ThreadPoolExecutor, wait, as_completed, FIRST_COMPLETED, ALL_COMPLETED, CancelledError, TimeoutError, Future


def square(x):
    return x * x


def slow(n, delay):
    time.sleep(delay)
    return n


with ThreadPoolExecutor(max_workers=3) as ex:
    futs = [ex.submit(square, i) for i in range(6)]
    print([f.result() for f in futs])
    print(list(ex.map(square, range(5))))
    print(list(ex.map(pow, [2, 3], [3, 2])))
    f = ex.submit(lambda: 1 / 0)
    try:
        f.result()
    except ZeroDivisionError as e:
        print("error", e)
    print(type(f.exception()).__name__, f.done(), f.cancelled(), f.running())
    order = []
    fs = [ex.submit(slow, 3, 0.09), ex.submit(slow, 1, 0.01), ex.submit(slow, 2, 0.05)]
    for fut in as_completed(fs):
        order.append(fut.result())
    print(order)
    done, not_done = wait([ex.submit(slow, "a", 0.01), ex.submit(slow, "b", 0.3)], return_when=FIRST_COMPLETED)
    print(sorted(f.result() for f in done), len(not_done))
    done, not_done = wait(not_done, timeout=2)
    print(len(done), len(not_done))
    seen = []
    cb = ex.submit(square, 7)
    cb.add_done_callback(lambda fut: seen.append(fut.result()))
    cb.result()
    time.sleep(0.02)
    print(seen)
    slowf = ex.submit(slow, 1, 0.2)
    try:
        slowf.result(timeout=0.01)
    except TimeoutError:
        print("timeout")
    print(slowf.result())

with ThreadPoolExecutor(max_workers=1, thread_name_prefix="job") as ex:
    names = ex.submit(lambda: threading.current_thread().name).result()
    print(names)
    blocker = ex.submit(slow, 0, 0.05)
    queued = ex.submit(square, 3)
    print(queued.cancel(), queued.cancelled(), blocker.cancel())
    try:
        queued.result()
    except CancelledError:
        print("cancelled")
ex = ThreadPoolExecutor(max_workers=2)
ex.shutdown()
try:
    ex.submit(square, 1)
except RuntimeError as e:
    print("RuntimeError", e)
fut = Future()
print(fut.done(), fut.running(), repr(fut).split(" ")[-1])
fut.set_result(5)
print(fut.result(), fut.done())
try:
    fut.set_result(6)
except Exception as e:
    print(type(e).__name__)
init = []
with ThreadPoolExecutor(max_workers=2, initializer=init.append, initargs=("ready",)) as ex:
    ex.submit(square, 1).result()
print(init[:1])
