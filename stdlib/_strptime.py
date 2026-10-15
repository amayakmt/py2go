"""_strptime: parse strings into time structures (the C locale)."""

import re

_DAYS = ["monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"]
_DAYS_ABBR = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"]
_MONTHS = ["", "january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"]
_MONTHS_ABBR = ["", "jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"]

_PATTERNS = {
    "d": r"(?P<d>3[0-1]|[1-2]\d|0[1-9]|[1-9]| [1-9])",
    "f": r"(?P<f>[0-9]{1,6})",
    "H": r"(?P<H>2[0-3]|[0-1]\d|\d)",
    "I": r"(?P<I>1[0-2]|0[1-9]|[1-9])",
    "G": r"(?P<G>\d\d\d\d)",
    "j": r"(?P<j>36[0-6]|3[0-5]\d|[1-2]\d\d|0[1-9]\d|00[1-9]|[1-9]\d|0[1-9]|[1-9])",
    "m": r"(?P<m>1[0-2]|0[1-9]|[1-9])",
    "M": r"(?P<M>[0-5]\d|\d)",
    "S": r"(?P<S>6[0-1]|[0-5]\d|\d)",
    "U": r"(?P<U>5[0-3]|[0-4]\d|\d)",
    "W": r"(?P<W>5[0-3]|[0-4]\d|\d)",
    "e": r"(?P<e>3[0-1]|[1-2]\d|0[1-9]|[1-9]| [1-9])",
    "w": r"(?P<w>[0-6])",
    "u": r"(?P<u>[1-7])",
    "V": r"(?P<V>5[0-3]|0[1-9]|[1-4]\d|\d)",
    "y": r"(?P<y>\d\d)",
    "Y": r"(?P<Y>\d\d\d\d)",
    "z": r"(?P<z>[+-]\d\d:?[0-5]\d(:?[0-5]\d(\.\d{1,6})?)?|(?-i:Z))",
    "A": r"(?P<A>monday|tuesday|wednesday|thursday|friday|saturday|sunday)",
    "a": r"(?P<a>mon|tue|wed|thu|fri|sat|sun)",
    "B": r"(?P<B>january|february|march|april|may|june|july|august|september|october|november|december)",
    "b": r"(?P<b>jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)",
    "p": r"(?P<p>am|pm)",
    "Z": r"(?P<Z>utc|gmt)",
    "%": "%",
}


def _expand(fmt):
    # compound directives
    out = ""
    i = 0
    n = len(fmt)
    while i < n:
        if fmt[i] == "%" and i + 1 < n:
            c = fmt[i + 1]
            if c == "c":
                out += "%a %b %d %H:%M:%S %Y"
            elif c == "x":
                out += "%m/%d/%y"
            elif c == "X":
                out += "%H:%M:%S"
            elif c == "T":
                out += "%H:%M:%S"
            elif c == "R":
                out += "%H:%M"
            elif c == "r":
                out += "%I:%M:%S %p"
            else:
                out += "%" + c
            i += 2
        else:
            out += fmt[i]
            i += 1
    return out


def _pattern(fmt):
    fmt = _expand(fmt)
    out = ""
    i = 0
    n = len(fmt)
    while i < n:
        ch = fmt[i]
        if ch == "%":
            if i + 1 >= n:
                raise ValueError("stray % in format '" + fmt + "'")
            c = fmt[i + 1]
            if c not in _PATTERNS:
                raise ValueError("'" + c + "' is a bad directive in format '" + fmt + "'")
            out += _PATTERNS[c]
            i += 2
        elif ch in " \t\n\r\x0b\x0c":
            while i < n and fmt[i] in " \t\n\r\x0b\x0c":
                i += 1
            out += r"\s+"
        else:
            out += re.escape(ch)
            i += 1
    return out


def _is_leap(y):
    return y % 4 == 0 and (y % 100 != 0 or y % 400 == 0)


def _ordinal(year, month, day):
    days = [31, 29 if _is_leap(year) else 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
    return sum(days[:month - 1]) + day


def _to_ordinal(year, month, day):
    y = year - 1
    return y * 365 + y // 4 - y // 100 + y // 400 + _ordinal(year, month, day)


def _from_ordinal(ordinal):
    n = ordinal - 1
    n400, n = divmod(n, 146097)
    year = n400 * 400 + 1
    n100, n = divmod(n, 36524)
    n4, n = divmod(n, 1461)
    n1, n = divmod(n, 365)
    year += n100 * 100 + n4 * 4 + n1
    if n1 == 4 or n100 == 4:
        return (year - 1, 12, 31)
    month = 1
    days = [31, 29 if _is_leap(year) else 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
    for d in days:
        if n < d:
            break
        n -= d
        month += 1
    return (year, month, n + 1)


def _calc_julian_from_weeks(year, week_of_year, day_of_week, first_weekday_arg):
    first_weekday = (_to_ordinal(year, 1, 1) + 6) % 7
    week_starts_Mon = first_weekday_arg == 0
    if not week_starts_Mon:
        first_weekday = (first_weekday + 1) % 7
        day_of_week = (day_of_week + 1) % 7
    week_0_length = (7 - first_weekday) % 7
    if week_of_year == 0:
        return 1 + day_of_week - first_weekday
    days_to_week = week_0_length + (7 * (week_of_year - 1))
    return 1 + days_to_week + day_of_week


def _strptime(data_string, format="%a %b %d %H:%M:%S %Y"):
    if not isinstance(data_string, str):
        raise TypeError("strptime() argument 1 must be str, not " + type(data_string).__name__)
    if not isinstance(format, str):
        raise TypeError("strptime() argument 2 must be str, not " + type(format).__name__)
    pat = _pattern(format)
    rx = re.compile(pat, re.IGNORECASE)
    found = rx.match(data_string)
    if found is None:
        raise ValueError("time data " + repr(data_string) + " does not match format " + repr(format))
    if len(data_string) != found.end():
        raise ValueError("unconverted data remains: " + data_string[found.end():])
    year = None
    month = None
    day = None
    hour = 0
    minute = 0
    second = 0
    fraction = 0
    tz = -1
    gmtoff = None
    tzname = None
    gmtoff_fraction = 0
    iso_year = None
    iso_week = None
    week_of_year = None
    week_of_year_start = None
    weekday = None
    julian = None
    ampm = None
    found_dict = found.groupdict()
    for key in found_dict.keys():
        v = found_dict[key]
        if v is None:
            continue
        if key == "y":
            year = int(v)
            if year <= 68:
                year += 2000
            else:
                year += 1900
        elif key == "Y":
            year = int(v)
        elif key == "G":
            iso_year = int(v)
        elif key == "m":
            month = int(v)
        elif key == "B":
            month = _MONTHS.index(v.lower())
        elif key == "b":
            month = _MONTHS_ABBR.index(v.lower())
        elif key == "d" or key == "e":
            day = int(v)
        elif key == "H":
            hour = int(v)
        elif key == "I":
            hour = int(v)
            ampm = ampm if ampm is not None else -1
        elif key == "M":
            minute = int(v)
        elif key == "S":
            second = int(v)
        elif key == "f":
            s = v
            s += "0" * (6 - len(s))
            fraction = int(s)
        elif key == "A":
            weekday = _DAYS.index(v.lower())
        elif key == "a":
            weekday = _DAYS_ABBR.index(v.lower())
        elif key == "w":
            weekday = int(v)
            if weekday == 0:
                weekday = 6
            else:
                weekday -= 1
        elif key == "u":
            weekday = int(v) - 1
        elif key == "j":
            julian = int(v)
        elif key == "U" or key == "W":
            week_of_year = int(v)
            week_of_year_start = 6 if key == "U" else 0
        elif key == "V":
            iso_week = int(v)
        elif key == "z":
            if v == "Z":
                gmtoff = 0
            else:
                z = v.replace(":", "")
                sign = -1 if z[0] == "-" else 1
                hh = int(z[1:3])
                mm = int(z[3:5])
                ss = int(z[5:7]) if len(z) > 5 else 0
                gmtoff = sign * (hh * 3600 + mm * 60 + ss)
        elif key == "Z":
            tz = 0
            tzname = v
        elif key == "p":
            ampm = 1 if v.lower() == "pm" else 0
    if ampm is not None and ampm != -1:
        pass
    # 12-hour clock
    if "I" in found_dict and found_dict["I"] is not None:
        pm = found_dict.get("p")
        if pm is not None and pm.lower() == "pm":
            if hour != 12:
                hour += 12
        elif pm is not None:
            if hour == 12:
                hour = 0
    leap_year_fix = False
    if year is None and month == 2 and day == 29:
        year = 1904
        leap_year_fix = True
    elif year is None:
        year = 1900
    if iso_year is not None:
        if julian is not None:
            raise ValueError("Day of the year directive '%j' is not compatible with ISO year directive '%G'. Use '%Y' instead.")
        elif iso_week is None or weekday is None:
            raise ValueError("ISO year directive '%G' must be used with the ISO week directive '%V' and a weekday directive ('%A', '%a', '%w', or '%u').")
    if julian is None and weekday is not None:
        if week_of_year is not None:
            week_starts_Mon = week_of_year_start == 0
            julian = _calc_julian_from_weeks(year, week_of_year, weekday, 0 if week_starts_Mon else 6)
            if julian <= 0:
                year -= 1
                yday = 366 if _is_leap(year) else 365
                julian += yday
    if iso_year is not None and julian is None:
        jan4 = _to_ordinal(iso_year, 1, 4)
        monday1 = jan4 - ((jan4 + 6) % 7)
        ordv = monday1 + (iso_week - 1) * 7 + weekday
        d3 = _from_ordinal(ordv)
        year = d3[0]
        month = d3[1]
        day = d3[2]
        julian = _ordinal(year, month, day)
    if julian is None:
        if month is None:
            month = 1
        if day is None:
            day = 1
        if month < 1 or month > 12:
            raise ValueError("month must be in 1..12")
        mdays = [31, 29 if _is_leap(year) else 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
        if day < 1 or day > mdays[month - 1]:
            raise ValueError("day is out of range for month" if False else "day is out of range for month")
        julian = _ordinal(year, month, day)
    else:
        datetime_result = _from_ordinal(_to_ordinal(year, 1, 1) + julian - 1)
        year = datetime_result[0]
        month = datetime_result[1]
        day = datetime_result[2]
    if weekday is None:
        weekday = (_to_ordinal(year, month, day) + 6) % 7
    if leap_year_fix:
        year = 1900
    return ((year, month, day, hour, minute, second, weekday, julian, tz, tzname, gmtoff), fraction, gmtoff_fraction)


def _strptime_time(data_string, format="%a %b %d %H:%M:%S %Y"):
    tt, fraction, gmtoff = _strptime(data_string, format)
    return tt[:9]
