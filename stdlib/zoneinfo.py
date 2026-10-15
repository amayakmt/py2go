"""zoneinfo: IANA time zone support."""
import datetime as _dt
from datetime import tzinfo as _tzinfo

__all__ = ["ZoneInfo", "reset_tzpath", "available_timezones", "TZPATH", "ZoneInfoNotFoundError", "InvalidTZPathWarning"]

TZPATH = ("/usr/share/zoneinfo", "/usr/lib/zoneinfo", "/usr/share/lib/zoneinfo", "/etc/zoneinfo")


class ZoneInfoNotFoundError(KeyError):
    pass


class InvalidTZPathWarning(RuntimeWarning):
    pass


def reset_tzpath(to=None):
    pass


def available_timezones():
    return set(_native("zoneinfo.available"))


_EPOCH = _dt.datetime(1970, 1, 1)


def _wall_seconds(dt):
    delta = dt.replace(tzinfo=None) - _EPOCH
    return delta.days * 86400 + delta.seconds


class ZoneInfo(_tzinfo):
    def __init__(self, key):
        if not isinstance(key, str):
            raise TypeError("ZoneInfo key must be a str")
        if key.startswith("/"):
            raise ValueError("ZoneInfo keys may not be absolute paths, got: " + key)
        if ".." in key.split("/") or "." in key.split("/") or "" in key.split("/"):
            raise ValueError("ZoneInfo keys must be normalized relative paths, got: " + key)
        if not _native("zoneinfo.exists", key):
            raise ZoneInfoNotFoundError("No time zone found with key " + key)
        self._key = key

    @classmethod
    def no_cache(cls, key):
        return cls(key)

    @classmethod
    def clear_cache(cls, *, only_keys=None):
        pass

    @property
    def key(self):
        return self._key

    def __repr__(self):
        return "zoneinfo.ZoneInfo(key=%r)" % self._key

    def __str__(self):
        return self._key

    def _local(self, dt):
        if dt is None:
            return None
        return _native("zoneinfo.at_local", self._key, _wall_seconds(dt), dt.fold)

    def utcoffset(self, dt):
        r = self._local(dt)
        if r is None:
            return None
        return _dt.timedelta(seconds=r[0])

    def dst(self, dt):
        r = self._local(dt)
        if r is None:
            return None
        return _dt.timedelta(seconds=r[1])

    def tzname(self, dt):
        r = self._local(dt)
        if r is None:
            return None
        return r[2]

    def fromutc(self, dt):
        if not isinstance(dt, _dt.datetime):
            raise TypeError("fromutc() requires a datetime argument")
        if dt.tzinfo is not self:
            raise ValueError("dt.tzinfo is not self")
        secs = _wall_seconds(dt)
        o, d, a = _native("zoneinfo.at_utc", self._key, secs)
        local = dt.replace(tzinfo=None) + _dt.timedelta(seconds=o)
        # the second occurrence of an ambiguous wall time has fold=1
        fold = 0
        first = _native("zoneinfo.at_local", self._key, secs + o, 0)
        if first[0] != o:
            fold = 1
        return local.replace(tzinfo=self, fold=fold)

    def __reduce__(self):
        return (ZoneInfo, (self._key,))
