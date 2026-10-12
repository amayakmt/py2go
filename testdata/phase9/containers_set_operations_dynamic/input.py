def combine(a, b):
    return a | b, a & b, a - b, a ^ b

print(combine({1, 2, 3}, {2, 3, 4}))
print([sorted(x) for x in combine({"a"}, {"a", "b"})])
s = set()
for x in [3, 1, 2, 3, 1]:
    s.add(x)
print(sorted(s), len(s), 2 in s)
s.discard(9)
s.remove(1)
print(sorted(s), s.issubset({2, 3, 4}), s.issuperset({2}), s.isdisjoint({7}))
t = {x * x for x in range(5)}
print(sorted(t), sorted(t | {100}), sorted(t.union([7, 8]), reverse=True))
u = set("hello")
print(sorted(u), len(u))
print({1, 2} < {1, 2, 3}, {1, 2} <= {1, 2}, {1} > set())
