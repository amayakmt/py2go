import threading
import time

lock = threading.Lock()
print(lock.locked(), lock.acquire(), lock.locked(), lock.acquire(False), lock.acquire(timeout=0.01))
lock.release()
try:
    lock.release()
except RuntimeError as e:
    print("RuntimeError", e)
with lock:
    print("inside", lock.locked())
print(lock.locked())
try:
    lock.acquire(False, 1)
except ValueError as e:
    print("ValueError", e)
try:
    lock.acquire(timeout=-5)
except ValueError as e:
    print("ValueError", e)

counter = 0
clock = threading.Lock()


def bump(n):
    global counter
    for _ in range(n):
        with clock:
            counter += 1


ts = [threading.Thread(target=bump, args=(1000,)) for _ in range(4)]
for t in ts:
    t.start()
for t in ts:
    t.join()
print(counter)

r = threading.RLock()
with r:
    with r:
        print("reentered")
r.acquire()
r.acquire()
r.release()
r.release()
try:
    r.release()
except RuntimeError as e:
    print("RuntimeError", e)

order = []
r2 = threading.RLock()


def holder():
    with r2:
        order.append("holder in")
        time.sleep(0.03)
        order.append("holder out")


h = threading.Thread(target=holder)
h.start()
time.sleep(0.01)
print(r2.acquire(False))
with r2:
    order.append("main in")
h.join()
print(order)

sem = threading.Semaphore(2)
active = []
peak = [0]
slock = threading.Lock()


def use(i):
    with sem:
        with slock:
            active.append(i)
            peak[0] = max(peak[0], len(active))
        time.sleep(0.01)
        with slock:
            active.remove(i)


ts = [threading.Thread(target=use, args=(i,)) for i in range(6)]
for t in ts:
    t.start()
for t in ts:
    t.join()
print(peak[0] <= 2, peak[0] >= 1, active)
bs = threading.BoundedSemaphore(1)
bs.acquire()
bs.release()
try:
    bs.release()
except ValueError as e:
    print("ValueError", e)
print(sem.acquire(blocking=False), sem.acquire(blocking=False), sem.acquire(blocking=False))
try:
    threading.Semaphore(-1)
except ValueError as e:
    print("ValueError", e)
