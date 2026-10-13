import time
t = time.gmtime(0)
print(t, t.tm_year, t.tm_mon, t.tm_mday, t.tm_hour, t.tm_wday, t.tm_yday, t.tm_isdst, tuple(t)[:3], len(t), t[0])
t2 = time.gmtime(1700000000)
print(t2, time.strftime("%Y-%m-%d %H:%M:%S", t2), time.strftime("%a %b %d %j %y %p %I", t2), time.strftime("%A %B %e|%%|%H%M", t2), time.asctime(t2))
print(time.gmtime(86400 * 365.25 * 30), time.mktime(time.localtime(1700000000)) == 1700000000.0)
a = time.time(); b = time.time()
print(type(a).__name__, b >= a, a > 1.6e9, type(time.time_ns()).__name__, time.perf_counter() >= 0, time.monotonic() <= time.monotonic(), type(time.perf_counter_ns()).__name__, time.process_time() >= 0)
start = time.perf_counter(); time.sleep(0.01); el = time.perf_counter() - start
print(el >= 0.009, el < 1.0)
for bad in (-1, "x"):
    try:
        time.sleep(bad)
    except (ValueError, TypeError) as e:
        print(type(e).__name__)
