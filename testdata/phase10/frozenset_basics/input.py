f = frozenset([3, 1, 2, 3])
g = frozenset("abca")
print(f, len(f), 2 in f, 5 in f, sorted(f), sorted(g), frozenset(), frozenset({1}), repr(frozenset([(1, 2)])), type(f).__name__, bool(frozenset()), bool(f))
print(f | {9}, f & {1, 2, 7}, f - {1}, f ^ {1, 8}, type(f | {9}).__name__, type({9} | f).__name__, {1, 2, 7} & f, f | frozenset([4]), f.union([5], [6]), f.intersection([1, 2]), f.difference([1]), f.symmetric_difference([1, 10]), f.copy() is f)
print(f == frozenset([1, 2, 3]), f == {1, 2, 3}, {1, 2, 3} == f, f != g, f <= f, f < f, f <= {1, 2, 3, 4}, f >= {1}, f > {1}, f.issubset([1, 2, 3, 4]), f.issuperset([1]), f.isdisjoint([7]), f.isdisjoint([1]))
d = {f: "a", frozenset([1, 2, 3]): "b"}
print(d, len(d), hash(f) == hash(frozenset([1, 2, 3])), hash(frozenset()) == hash(frozenset()), {frozenset([1]), frozenset([1]), frozenset([2])} == {frozenset([1]), frozenset([2])}, frozenset([frozenset([1]), frozenset()]))
s = {1, 2}
s |= frozenset([3])
t = frozenset([1])
t |= {2}
print(s, t, type(t).__name__, sorted(frozenset(range(5)) - frozenset(range(2, 4))), [x for x in f], list(map(lambda v: v * 2, sorted(f))), max(f), min(f), sum(f), any(f), all(f), sorted(f, reverse=True), len(frozenset("hello")))
for fn in (lambda: f.add(1), lambda: f.remove(1), lambda: f.discard(1), lambda: f.pop(), lambda: f.clear(), lambda: frozenset([[1]]), lambda: frozenset(5), lambda: f[0], lambda: f + f, lambda: f < 5, lambda: f | [1], lambda: f.union(5), lambda: hash(set())):
    try:
        print(fn())
    except (AttributeError, TypeError) as e:
        print(type(e).__name__, e)
