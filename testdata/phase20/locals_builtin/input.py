def f(a, b=2, *rest, **kw):
    total = a + b
    items = [1, 2]
    return sorted(locals().items(), key=lambda kv: kv[0])


print(f(1, 5, 9, x=3))


def g(name):
    greeting = "hi"
    return "{greeting}, {name}!".format(**locals())


print(g("ann"))
