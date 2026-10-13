"""time: clocks, sleeping and calendar fields."""
from collections import namedtuple

struct_time = namedtuple("time.struct_time", "tm_year tm_mon tm_mday tm_hour tm_min tm_sec tm_wday tm_yday tm_isdst")

timezone = 0
altzone = 0
daylight = 0
tzname = ("UTC", "UTC")


def time():
    return _native("time.time")


def time_ns():
    return _native("time.time_ns")


def sleep(secs):
    _native("time.sleep", secs)


def perf_counter():
    return _native("time.perf_counter")


def perf_counter_ns():
    return _native("time.perf_counter_ns")


def monotonic():
    return _native("time.perf_counter")


def monotonic_ns():
    return _native("time.perf_counter_ns")


def process_time():
    return _native("time.process_time")


def process_time_ns():
    return int(_native("time.process_time") * 1000000000)


def gmtime(secs=None):
    return struct_time(*_native("time.gmtime", secs))


def localtime(secs=None):
    return struct_time(*_native("time.localtime", secs))


def mktime(t):
    return _native("time.mktime", tuple(t))


def strftime(format, t=None):
    if t is None:
        t = localtime()
    return _native("time.strftime", format, tuple(t))


def strptime(data_string, format="%a %b %d %H:%M:%S %Y"):
    from _strptime import _strptime_time
    return struct_time(*_strptime_time(data_string, format))


def asctime(t=None):
    if t is None:
        t = localtime()
    return _native("time.strftime", "%a %b %e %H:%M:%S %Y", tuple(t))


def ctime(secs=None):
    return asctime(localtime(secs))
