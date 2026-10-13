import limits
from limits import clamp, MAX
total: int = 0
for v in (3, 12, 7):
    total += clamp(v)
print(total, MAX, limits.RATE, limits.scaled(3.0), limits.NAME, limits.TAGS, limits.LOOKUP["x"])
limits.TAGS.append("c")
print(len(limits.TAGS))
