import queue
import threading
import time

q = queue.Queue()
results = []


def worker():
    while True:
        item = q.get()
        if item is None:
            q.task_done()
            break
        results.append(item * item)
        q.task_done()


t = threading.Thread(target=worker)
t.start()
for i in range(6):
    q.put(i)
q.join()
q.put(None)
t.join()
print(results, q.empty(), q.qsize())
bq = queue.Queue(maxsize=2)
bq.put(1)
bq.put(2)
print(bq.full())
try:
    bq.put(3, block=False)
except queue.Full:
    print("full")
try:
    bq.put(3, timeout=0.01)
except queue.Full:
    print("full timeout")
print(bq.get(), bq.get())
try:
    bq.get_nowait()
except queue.Empty:
    print("empty")
try:
    bq.get(timeout=0.01)
except queue.Empty:
    print("empty timeout")
lq = queue.LifoQueue()
for i in range(3):
    lq.put(i)
print([lq.get() for _ in range(3)])
pq = queue.PriorityQueue()
for p in [(3, "c"), (1, "a"), (2, "b")]:
    pq.put(p)
print([pq.get() for _ in range(3)])
sq = queue.SimpleQueue()
sq.put("a")
sq.put("b")
print(sq.get(), sq.qsize(), sq.empty(), sq.get_nowait(), sq.empty())
try:
    sq.get_nowait()
except queue.Empty:
    print("simple empty")
try:
    q.task_done()
except ValueError as e:
    print("ValueError", e)
sq2 = queue.Queue()
sq2.put(1)
sq2.shutdown()
print(sq2.get())
try:
    sq2.get()
except queue.ShutDown:
    print("shut down")
try:
    sq2.put(2)
except queue.ShutDown:
    print("put shut down")
slow = queue.Queue(1)
out = []


def drain():
    time.sleep(0.02)
    while len(out) < 3:
        out.append(slow.get())


d = threading.Thread(target=drain)
d.start()
for i in range(3):
    slow.put(i)
d.join()
print(out)
