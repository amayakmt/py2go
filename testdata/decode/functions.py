@dec
@dec2(1)
def f(a, /, b=1, *args, k, j=2, **kw):
    nonlocal_ = 0


def outer():
    n = 0
    def inner():
        nonlocal n
        n += 1
    return lambda x, y=3: x + y


g(*xs, 1, *ys, key=2, **opts)
