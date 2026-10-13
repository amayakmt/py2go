import queue
import threading
import time
from concurrent.futures import ThreadPoolExecutor


class Stage(threading.Thread):
    def __init__(self, name, func, inq, outq):
        threading.Thread.__init__(self, name=name)
        self.func = func
        self.inq = inq
        self.outq = outq
        self.processed = 0

    def run(self):
        while True:
            item = self.inq.get()
            if item is None:
                if self.outq is not None:
                    self.outq.put(None)
                return
            result = self.func(item)
            self.processed += 1
            if self.outq is not None:
                self.outq.put(result)


class Collector:
    def __init__(self):
        self.lock = threading.Lock()
        self.items = []

    def add(self, item):
        with self.lock:
            self.items.append(item)


class RateLimiter:
    def __init__(self, per_second):
        self.interval = 1.0 / per_second
        self.cond = threading.Condition()
        self.next_time = time.monotonic()

    def wait(self):
        with self.cond:
            now = time.monotonic()
            delay = self.next_time - now
            if delay > 0:
                self.cond.wait(delay)
            self.next_time = max(now, self.next_time) + self.interval


def parse(line):
    k, v = line.split("=")
    return k.strip(), int(v)


def enrich(pair):
    return pair[0].upper(), pair[1] * 10


lines = ["a = 1", "b = 2", "c = 3", "d = 4", "e = 5"]
q1, q2, q3 = queue.Queue(), queue.Queue(), queue.Queue()
s1 = Stage("parse", parse, q1, q2)
s2 = Stage("enrich", enrich, q2, q3)
s1.start()
s2.start()
for line in lines:
    q1.put(line)
q1.put(None)
out = []
while True:
    item = q3.get()
    if item is None:
        break
    out.append(item)
s1.join()
s2.join()
print(out, s1.processed, s2.processed)

col = Collector()
rl = RateLimiter(200)
t0 = time.monotonic()


def job(n):
    rl.wait()
    col.add(n * n)
    return n


with ThreadPoolExecutor(max_workers=4) as pool:
    done = list(pool.map(job, range(10)))
print(done, sorted(col.items))
errors = []


def risky(n):
    if n % 3 == 0:
        raise ValueError("bad " + str(n))
    return n


with ThreadPoolExecutor(max_workers=3) as pool:
    futs = {pool.submit(risky, n): n for n in range(7)}
    for fut, n in futs.items():
        try:
            fut.result()
        except ValueError as e:
            errors.append((n, str(e)))
print(sorted(errors))
