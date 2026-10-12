s: set[int] = {3, 1, 2, 3, 1}
print(s, len(s), 2 in s, 9 in s, 9 not in s)
e: set[int] = set()
print(e, len(e), bool(e), bool(s))
s.add(10)
s.add(3)
print(s)
s.remove(1)
s.discard(1)
s.discard(99)
print(s)
print(sorted(s), min(s), max(s), sum(s))
t: set[tuple[int, int]] = {(1, 2), (2, 1), (1, 2)}
print(len(t), (1, 2) in t, (3, 3) in t)
flags: set[bool] = {True, False, True}
print(len(flags))
