def pair(a, b):
    return a, b

p = pair(1, "x")
print(p, type(p).__name__, len(p), p[0], p[1], p[-1])
a, b = p
print(a, b)
x, *rest = (1, 2, 3, 4)
print(x, rest)
*init, last = "abc"
print(init, last)
t = (1,) + (2, 3) * 2
print(t, t.count(2), t.index(3), 3 in t, t[1:3], t[::-1])
nested = ((1, 2), (3, (4, 5)))
(p1, p2), (p3, (p4, p5)) = nested
print(p1 + p2 + p3 + p4 + p5, nested[1][1][0])
print(tuple([1, 2]), tuple("ab"), tuple({1: 2}), tuple(range(3)), () == tuple())
print(sorted([(2, "b"), (1, "z"), (2, "a")]), max((1, 2), (1, 3)), (1, 2) < (1, 2, 0))
print(hash((1, 2)) == hash((1, 2)), {(1, 2): "a"}[(1, 2)])
