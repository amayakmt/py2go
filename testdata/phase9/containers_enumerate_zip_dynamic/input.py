names = ["ann", "bob", "cy"]
scores = [90, 85.5, 70]
for i, (n, s) in enumerate(zip(names, scores), start=1):
    print(i, n, s)
print(list(zip(names, scores, range(10))))
print(dict(zip(names, scores)))
print(list(zip(*[(1, 2), (3, 4), (5, 6)])))
print(list(enumerate("ab")), list(zip()), list(zip([1])))
total = 0
for a, b in zip([1, 2, 3], (10, 20, 30)):
    total += a * b
print(total)
print([x for x, in [(1,), (2,)]])
print(any(s > 90 for s in scores), all(isinstance(n, str) for n in names), any([]), all([]))
rev = list(reversed(list(enumerate(names))))
print(rev)
print(list(map(lambda p: p[0] + str(p[1]), zip(names, scores))))
print(list(filter(None, [0, 1, "", "a", None, [], [0]])))
