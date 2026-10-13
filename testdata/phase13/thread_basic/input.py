import threading
import time

results = {}


def work(name, n, scale=1):
    results[name] = sum(range(n)) * scale


threads = [threading.Thread(target=work, args=("t" + str(i), 100 * (i + 1)), kwargs={"scale": 2}) for i in range(4)]
print([t.name for t in threads])
print([t.is_alive() for t in threads], threads[0].ident)
for t in threads:
    t.start()
for t in threads:
    t.join()
print(sorted(results.items()))
print([t.is_alive() for t in threads], all(t.ident is not None for t in threads))
main = threading.current_thread()
print(main.name, main is threading.main_thread(), main.daemon, threading.active_count())


class Worker(threading.Thread):
    def __init__(self, label):
        threading.Thread.__init__(self, name="worker-" + label)
        self.label = label
        self.output = None

    def run(self):
        self.output = self.label.upper() + "!" + threading.current_thread().name


w = Worker("abc")
w.start()
w.join()
print(w.output, w.name, repr(w).startswith("<Worker(worker-abc, stopped"))
t = threading.Thread(target=lambda: None, name="named", daemon=True)
print(t.daemon, t.name)
try:
    t.join()
except RuntimeError as e:
    print("RuntimeError", e)
t.start()
t.join()
try:
    t.start()
except RuntimeError as e:
    print("RuntimeError", e)
t.name = "renamed"
print(t.name, t.is_alive())


def failing():
    raise ValueError("thread failure")


f = threading.Thread(target=failing)
f.start()
f.join()
print("survived", f.is_alive())
hooked = []
threading.excepthook = lambda args: hooked.append((args.exc_type.__name__, str(args.exc_value), args.thread.name))
g = threading.Thread(target=failing, name="g")
g.start()
g.join()
print(hooked)
fired = []
timer = threading.Timer(0.02, fired.append, args=["fired"])
timer.start()
timer2 = threading.Timer(5, fired.append, args=["never"])
timer2.start()
timer2.cancel()
timer.join()
timer2.join()
print(fired)
print(threading.get_ident() == threading.current_thread().ident, len(threading.enumerate()))
seen = []
th = threading.Thread(target=lambda: seen.append(threading.current_thread().name))
th.start()
th.join()
print(seen[0].startswith("Thread-"), seen[0].endswith("(<lambda>)"))
