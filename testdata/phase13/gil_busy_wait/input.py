import threading
import time

flag = []
count = [0]


def setter():
    time.sleep(0.02)
    flag.append(True)


threading.Thread(target=setter).start()
spins = 0
while not flag:
    spins += 1
print("flag seen", spins > 0)

stop = threading.Event()
ticks = [0]


def ticker():
    while not stop.is_set():
        ticks[0] += 1
        time.sleep(0.001)


t = threading.Thread(target=ticker)
t.start()
time.sleep(0.05)
stop.set()
t.join()
print(ticks[0] > 5)

total = [0]


def crunch(n):
    i = 0
    while i < n:
        i += 1
    total[0] += i


ts = [threading.Thread(target=crunch, args=(200000,)) for _ in range(3)]
start = time.time()
for x in ts:
    x.start()
for x in ts:
    x.join()
print(total[0])
