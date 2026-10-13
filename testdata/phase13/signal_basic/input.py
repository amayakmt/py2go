import os
import signal
import time

print(signal.SIGINT, repr(signal.SIGINT), int(signal.SIGINT), signal.SIGTERM.name, signal.SIGKILL.value)
print(signal.SIG_DFL, repr(signal.SIG_IGN), signal.getsignal(signal.SIGTERM))
print(signal.Signals(15), signal.Signals["SIGHUP"].value, signal.strsignal(2), signal.NSIG)
got = []


def handler(signum, frame):
    got.append(signum)


old = signal.signal(signal.SIGUSR1, handler)
print(old, signal.getsignal(signal.SIGUSR1) is handler)
os.kill(os.getpid(), signal.SIGUSR1)
signal.raise_signal(signal.SIGUSR1)
print(got)
print(signal.signal(signal.SIGUSR1, signal.SIG_IGN) is handler)
os.kill(os.getpid(), signal.SIGUSR1)
print(got)
signal.signal(signal.SIGALRM, handler)
print(signal.alarm(1))
t0 = time.time()
signal.pause()
print(got[-1], round(time.time() - t0))
print(signal.alarm(0))
try:
    signal.raise_signal(signal.SIGINT)
except KeyboardInterrupt:
    print("keyboard interrupt")


def boom(s, f):
    raise ValueError("from handler")


signal.signal(signal.SIGUSR2, boom)
try:
    os.kill(os.getpid(), signal.SIGUSR2)
    time.sleep(0.2)
except ValueError as e:
    print(e)
try:
    signal.signal(signal.SIGKILL, handler)
except OSError as e:
    print("OSError", e)
try:
    signal.signal(signal.SIGUSR1, "nope")
except TypeError as e:
    print("TypeError", e)
signal.signal(signal.SIGALRM, handler)
signal.setitimer(signal.ITIMER_REAL, 0.02)
time.sleep(0.3)
print(got[-1] == signal.SIGALRM, len(got))
signal.signal(signal.SIGINT, signal.default_int_handler)
print(signal.getsignal(signal.SIGINT) is signal.default_int_handler)
