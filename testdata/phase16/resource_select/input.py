import fcntl
import os
import resource
import select
import tempfile
import time

soft, hard = resource.getrlimit(resource.RLIMIT_NOFILE)
print(soft > 0, hard == resource.RLIM_INFINITY or hard >= soft, isinstance(resource.getpagesize(), int))
ru = resource.getrusage(resource.RUSAGE_SELF)
print(ru.ru_utime >= 0, ru.ru_maxrss > 0, len(ru))
resource.setrlimit(resource.RLIMIT_NOFILE, (soft, hard))
try:
    resource.getrlimit(9999)
except ValueError as e:
    print(e)

with tempfile.TemporaryFile() as f:
    fcntl.flock(f, fcntl.LOCK_EX | fcntl.LOCK_NB)
    fcntl.flock(f, fcntl.LOCK_UN)
    print(fcntl.fcntl(f, fcntl.F_GETFD) in (0, 1))

r, w = os.pipe()
print(select.select([r], [w], [], 0) == ([], [w], []))
os.write(w, b"x")
print(select.select([r], [], [], 1) == ([r], [], []))
t = time.time()
print(select.select([], [], [], 0.05), time.time() - t >= 0.04)
os.read(r, 1)
print(select.select([r], [], [], 0))
os.close(r)
os.close(w)
try:
    select.select([r], [], [], 0)
except OSError as e:
    print(type(e).__name__, e.errno)
