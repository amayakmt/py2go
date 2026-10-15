import calendar

print(calendar.isleap(2024), calendar.leapdays(2000, 2025), int(calendar.weekday(2024, 3, 5)), int(calendar.monthrange(2024, 2)[0]))
print(calendar.month(2024, 2))
print(calendar.month(2024, 3, w=3, l=2))
print(calendar.calendar(2024, w=2, c=4, m=4)[:600])
print(calendar.monthcalendar(2024, 2))
c = calendar.TextCalendar(calendar.SUNDAY)
print(c.formatmonth(2023, 10))
print(list(c.itermonthdays(2024, 2)), list(calendar.Calendar().itermonthdates(2024, 2))[:3])
print(calendar.day_name[0], calendar.month_abbr[3], list(calendar.day_abbr), calendar.month_name[12])
print(calendar.timegm((2024, 3, 5, 13, 45, 7)))
print(calendar.HTMLCalendar().formatmonth(2024, 2)[:300])
print(calendar.firstweekday(), calendar.Day.MONDAY, calendar.Month.MARCH, calendar.JANUARY)
try:
    calendar.month(2024, 13)
except calendar.IllegalMonthError as e:
    print(e)
try:
    calendar.weekday(2024, 1, 1) == 2
    calendar.setfirstweekday(9)
except calendar.IllegalWeekdayError as e:
    print(e)
print(calendar.weekheader(3))
calendar.prmonth(2024, 1)
