"""time: clocks, sleeping and calendar fields."""

CLOCK_REALTIME = 0
CLOCK_MONOTONIC = 1
CLOCK_PROCESS_CPUTIME_ID = 2
CLOCK_THREAD_CPUTIME_ID = 3
CLOCK_MONOTONIC_RAW = 4
CLOCK_BOOTTIME = 7
CLOCK_TAI = 11


class struct_time(tuple):
    n_sequence_fields = 9
    n_fields = 11
    n_unnamed_fields = 0

    def __init__(self, sequence, dict=None):
        items = list(sequence)
        if len(items) < 9 or len(items) > 11:
            if len(items) < 9:
                raise TypeError("time.struct_time() takes an at least 9-sequence (%d-sequence given)" % len(items))
            raise TypeError("time.struct_time() takes an at most 11-sequence (%d-sequence given)" % len(items))
        self.tm_year = items[0]
        self.tm_mon = items[1]
        self.tm_mday = items[2]
        self.tm_hour = items[3]
        self.tm_min = items[4]
        self.tm_sec = items[5]
        self.tm_wday = items[6]
        self.tm_yday = items[7]
        self.tm_isdst = items[8]
        self.tm_zone = items[9] if len(items) > 9 else None
        self.tm_gmtoff = items[10] if len(items) > 10 else None
        self._set(tuple(items[:9]))  # py2go: the tuple value of the tuple subclass

    def _t(self):
        return (self.tm_year, self.tm_mon, self.tm_mday, self.tm_hour, self.tm_min, self.tm_sec,
                self.tm_wday, self.tm_yday, self.tm_isdst)

    def _full(self):
        return self._t() + (self.tm_zone, self.tm_gmtoff)

    def __getitem__(self, i):
        return self._t()[i]

    def __iter__(self):
        return iter(self._t())

    def __len__(self):
        return 9

    def __eq__(self, other):
        if isinstance(other, struct_time):
            return self._t() == other._t()
        return self._t() == other

    def __lt__(self, other):
        return self._t() < tuple(other)

    def __le__(self, other):
        return self._t() <= tuple(other)

    def __gt__(self, other):
        return self._t() > tuple(other)

    def __ge__(self, other):
        return self._t() >= tuple(other)

    def __hash__(self):
        return hash(self._t())

    def __repr__(self):
        return ("time.struct_time(tm_year=%r, tm_mon=%r, tm_mday=%r, tm_hour=%r, tm_min=%r, tm_sec=%r, "
                "tm_wday=%r, tm_yday=%r, tm_isdst=%r)" % self._t())

    def __reduce__(self):
        return (struct_time, (self._t(),))


def _tz():
    return _native("time.tzinfo")


timezone = _tz()[0]
altzone = _tz()[1]
daylight = 1 if _tz()[2] else 0
tzname = (_tz()[3], _tz()[4])


def tzset():
    global timezone, altzone, daylight, tzname
    _native("time.tzset")
    info = _tz()
    timezone = info[0]
    altzone = info[1]
    daylight = 1 if info[2] else 0
    tzname = (info[3], info[4])


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
    return _native("time.process_time_ns")


def thread_time():
    return _native("time.process_time")


def thread_time_ns():
    return _native("time.process_time_ns")


def clock_gettime(clk_id):
    if clk_id == CLOCK_REALTIME:
        return _native("time.time")
    if clk_id == CLOCK_MONOTONIC or clk_id == CLOCK_MONOTONIC_RAW or clk_id == CLOCK_BOOTTIME:
        return _native("time.perf_counter")
    if clk_id == CLOCK_PROCESS_CPUTIME_ID or clk_id == CLOCK_THREAD_CPUTIME_ID:
        return _native("time.process_time")
    raise OSError(22, "Invalid argument")


def clock_gettime_ns(clk_id):
    if clk_id == CLOCK_REALTIME:
        return _native("time.time_ns")
    return int(clock_gettime(clk_id) * 1000000000)


def clock_getres(clk_id):
    return 1e-09


class _clock_info:
    def __init__(self, implementation, monotonic, adjustable, resolution):
        self.implementation = implementation
        self.monotonic = monotonic
        self.adjustable = adjustable
        self.resolution = resolution

    def __repr__(self):
        return ("namespace(implementation=%r, monotonic=%r, adjustable=%r, resolution=%r)"
                % (self.implementation, self.monotonic, self.adjustable, self.resolution))


def get_clock_info(name):
    if name == "time":
        return _clock_info("clock_gettime(CLOCK_REALTIME)", False, True, 1e-09)
    if name == "monotonic":
        return _clock_info("clock_gettime(CLOCK_MONOTONIC)", True, False, 1e-09)
    if name == "perf_counter":
        return _clock_info("clock_gettime(CLOCK_MONOTONIC)", True, False, 1e-09)
    if name == "process_time":
        return _clock_info("clock_gettime(CLOCK_PROCESS_CPUTIME_ID)", True, False, 1e-09)
    if name == "thread_time":
        return _clock_info("clock_gettime(CLOCK_THREAD_CPUTIME_ID)", True, False, 1e-09)
    raise ValueError("unknown clock")


def _wrap(fields):
    return struct_time(fields)


def gmtime(secs=None):
    return struct_time(_native("time.gmtime", secs))


def localtime(secs=None):
    return struct_time(_native("time.localtime", secs))


def mktime(t):
    return _native("time.mktime", _tuple(t))


def _tuple(t):
    if isinstance(t, struct_time):
        return t._full()
    if not isinstance(t, tuple):
        raise TypeError("Tuple or struct_time argument required")
    return t


def strftime(format, t=None):
    if not isinstance(format, str):
        raise TypeError("strftime() argument 1 must be str, not " + type(format).__name__)
    if t is None:
        t = localtime()
    return _native("time.strftime", format, _tuple(t))


def strptime(data_string, format="%a %b %d %H:%M:%S %Y"):
    from _strptime import _strptime_time
    return struct_time(_strptime_time(data_string, format))


def asctime(t=None):
    if t is None:
        t = localtime()
    return _native("time.strftime", "%a %b %e %H:%M:%S %Y", _tuple(t))


def ctime(secs=None):
    return asctime(localtime(secs))
