import time

fmts = [
    ("2024-03-15 13:45:30", "%Y-%m-%d %H:%M:%S"), ("15/03/24", "%d/%m/%y"), ("Fri Mar 15 13:45:30 2024", "%c"), ("March 5, 1999", "%B %d, %Y"),
    ("Dec 25 2020 11:59 PM", "%b %d %Y %I:%M %p"), ("12:00 AM", "%I:%M %p"), ("12:00 PM", "%I:%M %p"), ("2021 100", "%Y %j"), ("2020-02-29", "%Y-%m-%d"),
    ("Mon Jan  1 00:00:00 2018", "%a %b %d %H:%M:%S %Y"), ("1970-01-01T00:00:00+0530", "%Y-%m-%dT%H:%M:%S%z"), ("23:59:60", "%H:%M:%S"), ("05/06", "%m/%d"),
    ("2022 10 3", "%Y %U %w"), ("2022 10 1", "%Y %W %u"), ("  7  ", " %d "), ("1 2 3", "%d %m %y"), ("100%", "%Y%%") if False else ("99%", "%y%%"),
    ("2024-W11-5", "%G-W%V-%u"), ("07:08", "%R"), ("07:08:09 PM", "%r"), (" 5", "%e"), ("Saturday", "%A"), ("sat", "%a"), ("07:08:09", "%T"), ("UTC", "%Z"),
]
for text, fmt in fmts:
    try:
        t = time.strptime(text, fmt)
        print(repr(text), fmt, tuple(t))
    except ValueError as e:
        print(repr(text), fmt, "ValueError", e)
for text, fmt in [("2024-13-01", "%Y-%m-%d"), ("2024-02-30", "%Y-%m-%d"), ("2024-03-15x", "%Y-%m-%d"), ("abc", "%Y"), ("24", "%Q"), ("2024", "%Y%"), ("12", "%H:%M"), ("Febuary 3", "%B %d"), ("2023-02-29", "%Y-%m-%d")]:
    try:
        print(tuple(time.strptime(text, fmt)))
    except ValueError as e:
        print("ValueError", e)
print(time.strptime("2001-02-03", "%Y-%m-%d").tm_year, time.strptime("2001-02-03", "%Y-%m-%d").tm_yday, time.strptime("2001-02-03", "%Y-%m-%d").tm_wday)
