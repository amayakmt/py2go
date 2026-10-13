import threading
import time

ev = threading.Event()
log = []


def waiter(i):
    log.append("wait" + str(i))
    ev.wait()
    log.append("go" + str(i))


ts = [threading.Thread(target=waiter, args=(i,)) for i in range(3)]
for t in ts:
    t.start()
time.sleep(0.03)
print(sorted(log), ev.is_set())
ev.set()
for t in ts:
    t.join()
print(sorted(log), ev.is_set())
ev.clear()
print(ev.wait(0.01), ev.is_set())

cond = threading.Condition()
box = []
got = []


def consumer(name):
    with cond:
        while not box:
            cond.wait()
        got.append((name, box.pop(0)))


cs = [threading.Thread(target=consumer, args=("c" + str(i),)) for i in range(2)]
for c in cs:
    c.start()
time.sleep(0.03)
with cond:
    box.append("x")
    box.append("y")
    cond.notify_all()
for c in cs:
    c.join()
print(sorted(got))
with cond:
    print(cond.wait(0.01), cond.wait_for(lambda: False, 0.01))
try:
    cond.notify()
except RuntimeError as e:
    print("RuntimeError", e)

barrier = threading.Barrier(3)
idx = []


def party():
    idx.append(barrier.wait())


ps = [threading.Thread(target=party) for _ in range(3)]
for p in ps:
    p.start()
for p in ps:
    p.join()
print(sorted(idx), barrier.parties, barrier.n_waiting, barrier.broken)
b2 = threading.Barrier(2, timeout=0.02)
try:
    b2.wait()
except threading.BrokenBarrierError:
    print("broken", b2.broken)
b2.reset()
print(b2.broken)

local = threading.local()
local.value = "main"
seen = []


def reader(v):
    seen.append(hasattr(local, "value"))
    local.value = v
    time.sleep(0.01)
    seen.append(local.value)


rs = [threading.Thread(target=reader, args=("t" + str(i),)) for i in range(2)]
for r in rs:
    r.start()
for r in rs:
    r.join()
print(sorted(map(str, seen)), local.value)
