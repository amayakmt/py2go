x = 1 if y else 2
(z := 5)
with open(f) as fh, g():
    pass


def gen():
    yield 1
    v = yield
    yield from other()


match p:
    case 1 | 2:
        pass
    case [a, *rest] if a:
        pass
    case {"k": v, **kw}:
        pass
    case Point(1, y=q) as pt:
        pass
    case None:
        pass
    case _:
        pass
