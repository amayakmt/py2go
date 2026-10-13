import sched
import time

s = sched.scheduler(time.monotonic, time.sleep)
out = []
s.enter(0.03, 1, out.append, ("c",))
s.enter(0.01, 1, out.append, ("a",))
e = s.enter(0.02, 1, out.append, ("b",))
s.enter(0.01, 0, out.append, ("first",))
s.enter(0.04, 1, lambda **kw: out.append(kw), kwargs={"k": 1})
print(len(s.queue), s.empty(), s.queue[0].priority)
s.cancel(e)
try:
    s.cancel(e)
except ValueError:
    print("not queued")
s.run()
print(out, s.empty())
s.enterabs(time.monotonic() + 1, 1, out.append, ("later",))
print(s.run(blocking=False) > 0.5, len(out))
clock = [0.0]
fake = sched.scheduler(lambda: clock[0], lambda d: clock.__setitem__(0, clock[0] + d))
log = []
fake.enter(5, 1, log.append, ("five",))
fake.enter(2, 1, log.append, ("two",))
fake.enter(2, 0, log.append, ("two-first",))
fake.run()
print(log, clock[0])
