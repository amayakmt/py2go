import datetime
from datetime import date, time, timedelta, datetime as dt, timezone

d = date(2024, 2, 29)
print(d, repr(d), d.weekday(), d.isoweekday(), d.isocalendar(), d.toordinal())
print(d + timedelta(days=366), d - date(2000, 1, 1))
print(date.fromisoformat("2023-12-31"), date.fromordinal(738000))
t = dt(2024, 3, 5, 13, 45, 7, 123456)
print(t, repr(t), t.isoformat(), t.isoformat(sep=" ", timespec="milliseconds"))
print(t.strftime("%Y-%m-%d %H:%M:%S.%f %a %b %j"))
print(t.replace(year=2000, tzinfo=timezone.utc))
u = dt(2024, 3, 5, 13, 45, tzinfo=timezone(timedelta(hours=5, minutes=30)))
print(u, u.utcoffset(), u.astimezone(timezone.utc), u.timestamp())
print(timedelta(days=1, seconds=3661, microseconds=5), timedelta(hours=-5), timedelta(1.5) * 3)
print(dt.fromisoformat("2024-01-02T03:04:05+01:00"))
print(dt.strptime("2021-07-04 10:11", "%Y-%m-%d %H:%M"))
print(t < dt(2025, 1, 1), t == dt(2024, 3, 5, 13, 45, 7, 123456), sorted([d, date(2000, 1, 1)]))
print(time(1, 2, 3), time.fromisoformat("04:05:06"), datetime.MINYEAR, datetime.MAXYEAR)
print(dt.fromtimestamp(0, timezone.utc), dt.utcfromtimestamp(86400) if False else "")
try:
    date(2023, 2, 30)
except ValueError as e:
    print(e)
print(format(t, "%d/%m/%y"), f"{d:%B %d, %Y}", str(timedelta.max), timedelta.resolution)
