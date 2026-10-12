def f(a, b=2, *rest, c, d=4, **extra):
    return (a, b, rest, c, d, sorted(extra.items()))

args = (1, 2, 3, 4)
kw = {"c": 5, "e": 6}
print(f(*args, **kw))
print(f(1, c=3), f(a=1, c=2, z=0), f(*[1], c=0, **{"d": 9}))
def apply(fn, *args, **kwargs):
    return fn(*args, **kwargs)
print(apply(f, 1, c=2), apply(max, 3, 9, 4), apply(sorted, [3, 1, 2], reverse=True), apply(print, "x", "y", sep="-"))
for bad in (lambda: f(), lambda: f(1), lambda: f(1, 2, 3, 4, 5, c=1, c2=2, a=3), lambda: f(c=1)):
    try:
        bad()
    except TypeError as e:
        print("TypeError", e)
def first(*items, default=None):
    return items[0] if items else default
print(first(), first(1, 2), first(default="d"), first(*"xyz"))
print(*[1, 2, 3], sep=", ", end="!\n")
print(dict(a=1, **{"b": 2}), list(zip(*[[1, 2], [3, 4]])), [*range(3), *"ab"], {**{"a": 1}, "b": 2})
