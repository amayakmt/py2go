def attempt(f, *args):
    try:
        return f(*args)
    except Exception as e:
        return type(e).__name__ + ": " + str(e)

print(attempt(lambda a, b: a / b, 1, 0))
print(attempt(lambda a, b: a + b, 1, "x"))
print(attempt(lambda a, b: a + b, [1], (2,)))
print(attempt(lambda d, k: d[k], {"a": 1}, "b"))
print(attempt(lambda l, i: l[i], [1, 2], 5))
print(attempt(lambda s: s.nope, "text"))
print(attempt(lambda n: int(n), "abc"))
print(attempt(lambda x: x(), 5))
print(attempt(lambda x: len(x), 5))
print(attempt(lambda x: -x, "s"))
print(attempt(lambda x: x < 1, "s"))
print(attempt(lambda l: l.index(9), [1]))
print(attempt(lambda a, b: a % b, 5, 0))
print(attempt(lambda: [][0]))
print(attempt(lambda t: t[0] if t else 1 // 0, ()))
print(attempt(lambda x: sorted(x), [3, "a"]))
print(attempt(lambda x: max(x), []))
print(attempt(lambda x: next(x), iter([])))
