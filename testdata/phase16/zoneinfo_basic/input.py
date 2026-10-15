from datetime import datetime, timedelta, timezone
from zoneinfo import ZoneInfo, ZoneInfoNotFoundError, available_timezones

ny = ZoneInfo("America/New_York")
print(ny, repr(ny), ny.key)
d = datetime(2024, 3, 9, 12, 0, tzinfo=ny)
print(d, d.utcoffset(), d.dst(), d.tzname())
d2 = datetime(2024, 3, 11, 12, 0, tzinfo=ny)
print(d2, d2.utcoffset(), d2.dst(), d2.tzname())
print(d2 - d, d2.astimezone(timezone.utc), d.astimezone(ZoneInfo("Asia/Tokyo")), d.astimezone(ZoneInfo("Europe/London")))
u = datetime(2024, 11, 3, 5, 30, tzinfo=timezone.utc)
a = u.astimezone(ny)
b = (u + timedelta(hours=1)).astimezone(ny)
print(a, a.fold, a.tzname(), b, b.fold, b.tzname())
print(datetime(2024, 11, 3, 1, 30, tzinfo=ny).utcoffset(), datetime(2024, 11, 3, 1, 30, fold=1, tzinfo=ny).utcoffset())
print(datetime(2024, 3, 10, 2, 30, tzinfo=ny).utcoffset(), datetime(2024, 3, 10, 2, 30, fold=1, tzinfo=ny).utcoffset())
print(datetime(2024, 6, 1, tzinfo=ZoneInfo("UTC")), ZoneInfo("Asia/Kolkata").utcoffset(datetime(2024, 1, 1)), ZoneInfo("Australia/Lord_Howe").dst(datetime(2024, 1, 1)))
print(datetime(2024, 1, 1, tzinfo=ZoneInfo("Europe/Paris")).timestamp(), datetime(2024, 7, 1, tzinfo=ZoneInfo("Europe/Paris")).isoformat())
print(datetime.fromtimestamp(1700000000, ZoneInfo("Asia/Tokyo")), datetime.now(ZoneInfo("UTC")).tzinfo)
try:
    ZoneInfo("Not/AZone")
except ZoneInfoNotFoundError as e:
    print(type(e).__name__, e)
try:
    ZoneInfo("/etc/passwd")
except ValueError as e:
    print(e)
print("Europe/Berlin" in available_timezones(), "UTC" in available_timezones(), len(available_timezones()) > 100)
