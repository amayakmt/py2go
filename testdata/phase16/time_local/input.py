import os
import time
import datetime

os.environ["TZ"] = "UTC"
time.tzset()
print(time.timezone, time.altzone, time.daylight, time.tzname)
t = time.gmtime(0)
print(t, t.tm_year, t[0], len(t), t.tm_zone, t.tm_gmtoff, tuple(t), t == (1970, 1, 1, 0, 0, 0, 3, 1, 0))
print(time.gmtime(1700000000.9), time.gmtime(-1))
print(time.asctime(t), time.ctime(0), time.asctime(time.gmtime(86400 * 40)))
print(time.strftime("%Y-%m-%d %H:%M:%S %a %A %b %B %j %U %W %w %u %y %Z %z %p %I %%", time.gmtime(1700000000)))
print(time.strftime("%c|%x|%X|%D|%F|%T|%R|%r|%e|%k|%l|%C|%G-%V|%h|%n|%t|%s", time.gmtime(1700000000)))
print(time.strftime("%-d/%-m/%y %_d %^a %#Z %5Y", time.gmtime(1700000000 + 86400 * 5)))
print(time.mktime(time.gmtime(1700000000)), time.mktime((2024, 2, 30, 12, 0, 0, 0, 0, -1)), time.localtime(1700000000).tm_isdst)
for bad in [lambda: time.strftime("%Y", (2024, 13, 1, 0, 0, 0, 0, 1, 0)), lambda: time.strftime("%Y", (2024, 1, 1)),
            lambda: time.gmtime("x"), lambda: time.strftime(5), lambda: time.mktime((1,)), lambda: time.struct_time((1, 2))]:
    try:
        bad()
    except (ValueError, TypeError, OverflowError) as e:
        print(type(e).__name__, e)

os.environ["TZ"] = "America/New_York"
time.tzset()
print(time.timezone, time.altzone, time.daylight, time.tzname)
w = time.localtime(1700000000)
s = time.localtime(1690000000)
print(w.tm_hour, w.tm_zone, w.tm_gmtoff, w.tm_isdst, s.tm_hour, s.tm_zone, s.tm_gmtoff, s.tm_isdst)
print(time.mktime(w), time.mktime(s), time.strftime("%H:%M %Z %z", w), time.strftime("%H:%M %Z %z", s))
print(time.ctime(1700000000))
os.environ["TZ"] = "Asia/Kolkata"
time.tzset()
print(time.timezone, time.daylight, time.tzname, time.localtime(0).tm_hour, time.localtime(0).tm_min)
os.environ["TZ"] = "EST5EDT"
time.tzset()
print(time.timezone, time.tzname)
os.environ["TZ"] = "UTC"
time.tzset()

print(time.strptime("2024-03-05 13:45:07", "%Y-%m-%d %H:%M:%S"))
print(time.strptime("Tue Mar  5 2024", "%a %b %d %Y").tm_yday)
print(type(time.time()).__name__, time.time_ns() > 0, time.monotonic() <= time.monotonic(), time.perf_counter_ns() > 0, time.process_time() >= 0)
print(time.get_clock_info("monotonic").monotonic, time.clock_gettime(time.CLOCK_REALTIME) > 1e9)
a = time.time()
time.sleep(0.02)
print(0.015 < time.time() - a < 0.5)
try:
    time.sleep(-1)
except ValueError as e:
    print(e)

# datetime with local time
os.environ["TZ"] = "Europe/Paris"
time.tzset()
d = datetime.datetime.fromtimestamp(1700000000)
print(d, d.astimezone(), d.astimezone().tzname(), d.astimezone().utcoffset())
print(datetime.datetime.fromtimestamp(1690000000).astimezone(), datetime.datetime(2024, 7, 1, 12).astimezone().isoformat())
print(datetime.datetime.fromtimestamp(0, datetime.timezone.utc).astimezone().strftime("%Y-%m-%d %H:%M:%S %Z"))
print(datetime.datetime(2024, 1, 15, 8, 30).timestamp(), datetime.datetime(2024, 7, 15, 8, 30).timestamp())
print(datetime.datetime.utcfromtimestamp(0), datetime.datetime.now().year >= 2024, datetime.date.today().year >= 2024)
print(datetime.datetime.now(datetime.timezone.utc).tzinfo, datetime.datetime.now(datetime.timezone(datetime.timedelta(hours=3))).utcoffset())
