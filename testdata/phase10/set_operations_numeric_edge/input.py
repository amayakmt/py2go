a = {1, 2.0, True, 3 + 0j, 2 ** 70, float(2 ** 70), "x", b"x", (1, 2), (1.0, 2), frozenset({1}), frozenset({1.0}), None}
print(len(a), sorted(map(repr, a)))
b = {1.0, 2, 4, 2 ** 70, "x", None}
print(sorted(map(repr, a & b)), sorted(map(repr, a - b)), sorted(map(repr, a ^ b)), len(a | b), a.isdisjoint({99}), b <= a, {1} <= {True}, {2 ** 70} <= a)
s = set()
for x in (0, 0.0, -0.0, False, 0j, 1, True, 1.0, 1 + 0j):
    s.add(x)
print(s, len(s))
s.discard(0.0)
s.add(2.5)
print(s, 1 in s, True in s, 0 in s, 2.5 in s, sorted(s))
t = {(2 ** 65 + k) % 7 for k in range(30)}
print(sorted(t), len(t), {x for x in range(10) if x % 3 == 0} | {x * 2 for x in range(3)}, {(i, j) for i in range(3) for j in range(i)})
u = {float("nan")}
u.add(float("nan"))
print(len(u), float("inf") in {float("inf")}, -0.0 in {0}, {0: "a"}[-0.0], {1: "a"}[True])
for f in (lambda: {1: 1}[2], lambda: set(5), lambda: {1} | [1], lambda: {1}.remove(2), lambda: {1} < [1], lambda: {1, 2}.pop() + {3}.pop() if False else {}.pop()):
    try:
        print(f())
    except (TypeError, KeyError) as e:
        print(type(e).__name__, e)
