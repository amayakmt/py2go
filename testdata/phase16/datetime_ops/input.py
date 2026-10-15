from datetime import date, time, timedelta, datetime, timezone, tzinfo
import datetime as dtm


class Fixed(tzinfo):
    def __init__(self, hours, name):
        self.off = timedelta(hours=hours)
        self.name = name

    def utcoffset(self, d):
        return self.off

    def dst(self, d):
        return timedelta(0)

    def tzname(self, d):
        return self.name


td = timedelta(days=2, hours=3, minutes=4, seconds=5, milliseconds=6, microseconds=7)
print(td, repr(td), td.total_seconds(), td.days, td.seconds, td.microseconds)
print(td * 2, td / 2, td // 3, td % timedelta(hours=5), divmod(td, timedelta(days=1)), -td, abs(-td))
print(td / timedelta(hours=1), td // timedelta(hours=1), td * 1.5, bool(timedelta(0)))
print(timedelta(weeks=1), timedelta(seconds=-1), timedelta(microseconds=-1), timedelta(days=-1, hours=23))
print(sorted([timedelta(1), timedelta(0), timedelta(-1)]), max(timedelta(1), timedelta(2)), hash(timedelta(1)) == hash(timedelta(hours=24)))
print(timedelta.min, timedelta.max, timedelta.resolution)

d = date(2024, 1, 31)
print(d.replace(month=2, day=29), d + timedelta(days=30), d - timedelta(days=31), d.ctime(), d.timetuple())
print(d.isocalendar().week, d.isocalendar()[0], d.isoformat(), d.strftime("%A %d %B %Y, week %U/%W, day %j"))
print(date.today() >= date(2024, 1, 1), date.max, date.min, date.resolution)
print(date(2024, 12, 30).isocalendar(), date(2021, 1, 3).isocalendar(), date.fromisocalendar(2020, 53, 7))
print(date.fromisoformat("20240131"), date.fromisoformat("2024-W05-3"))
print(d == date(2024, 1, 31), d != date(2024, 1, 30), d < date(2024, 2, 1), hash(d) == hash(date(2024, 1, 31)))
print({d: 1}[date(2024, 1, 31)], len({d, date(2024, 1, 31)}))

t = time(23, 59, 58, 999999)
print(t, repr(t), t.isoformat(timespec="seconds"), t.replace(hour=1), t.strftime("%I:%M %p"), time.min, time.max)
print(time(1, 2, tzinfo=timezone.utc), repr(time(1, 2, 3, tzinfo=timezone(timedelta(hours=-3)))), time(0, 0, fold=1).fold)

z = timezone(timedelta(hours=-5, minutes=-30), "EST5")
print(z, repr(z), z.tzname(None), z.utcoffset(None), timezone.utc, repr(timezone.utc), timezone(timedelta(hours=1)))
n = datetime(2024, 6, 15, 12, 30, 45, 500)
a = n.replace(tzinfo=z)
print(a, a.astimezone(timezone.utc), a.astimezone(timezone(timedelta(hours=9))), a.tzname(), a.utcoffset(), a.dst())
print(a.timetuple(), a.utctimetuple(), a.date(), a.time(), a.timetz())
print(n.isocalendar(), n.weekday(), n.toordinal(), n.timestamp() if False else "", datetime.fromordinal(739000))
f = n.replace(tzinfo=Fixed(2, "EET"))
print(f, f.tzname(), f.astimezone(timezone.utc), f.isoformat(), f.strftime("%z %Z"))
print(n - datetime(2024, 1, 1), n + timedelta(days=200, hours=13), n - timedelta(microseconds=501))
print(datetime.combine(date(2020, 2, 3), time(4, 5, 6)), datetime.min, datetime.max, datetime.resolution)
print(datetime.fromtimestamp(1700000000, timezone.utc), datetime.fromtimestamp(1700000000.5, timezone(timedelta(hours=1))))
print(datetime.fromisoformat("2024-02-29 23:59:59.123456"), datetime.fromisoformat("2024-02-29T23:59Z"), datetime.fromisoformat("20240229T235959"))
print(datetime.strptime("Tue, 05 Mar 2024 01:02:03 +0530", "%a, %d %b %Y %H:%M:%S %z"))
print(n.strftime("%y%m%d %H%M%S %f %%"), format(n, "%Y/%m/%d"), str(n), n.isoformat(timespec="hours"))
print(sorted([datetime(2024, 1, 2), datetime(2023, 1, 2), datetime(2025, 1, 2)]))
print(datetime(2024, 1, 1, tzinfo=timezone.utc) == datetime(2024, 1, 1, 1, tzinfo=timezone(timedelta(hours=1))))
print(isinstance(n, date), isinstance(d, datetime), type(n).__name__, dtm.UTC is timezone.utc)

for bad in [lambda: date(2024, 13, 1), lambda: date(0, 1, 1), lambda: time(24), lambda: time(1, 60), lambda: datetime(2024, 1, 1, 0, 0, 0, 1000000),
            lambda: timedelta(days=1000000000), lambda: timezone(timedelta(days=1)), lambda: date.fromisoformat("2024-1-1"),
            lambda: datetime(2024, 1, 1) - date(2024, 1, 1), lambda: d + 1, lambda: datetime(2024, 1, 1) < date(2024, 1, 1),
            lambda: date(2024, 1, 1) + date(2024, 1, 1), lambda: timedelta(1) / 0, lambda: datetime.strptime("x", "%Y"),
            lambda: datetime(2024, 1, 1, tzinfo=timezone.utc) - datetime(2024, 1, 1), lambda: date.max + timedelta(1)]:
    try:
        bad()
    except Exception as e:
        print(type(e).__name__, e)
