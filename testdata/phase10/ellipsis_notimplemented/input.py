e = Ellipsis
print(e, ..., repr(...), type(...).__name__, ... is Ellipsis, ... == Ellipsis, bool(...), hash(...) == hash(Ellipsis), [...], (..., 1), {...: 1}[Ellipsis], ... in [Ellipsis], str(...), f"{...}")
n = NotImplemented
print(n, repr(NotImplemented), type(n).__name__, n is NotImplemented, n == NotImplemented, bool(n) if False else 0, [n], str(n))
class V:
    def __init__(self, v):
        self.v = v
    def __eq__(self, o):
        if not isinstance(o, V):
            return NotImplemented
        return self.v == o.v
    def __lt__(self, o):
        if not isinstance(o, V):
            return NotImplemented
        return self.v < o.v
    def __add__(self, o):
        if isinstance(o, int):
            return V(self.v + o)
        return NotImplemented
    def __radd__(self, o):
        return self + o
    def __repr__(self):
        return f"V({self.v})"
print(V(1) == V(1), V(1) == 1, 1 == V(1), V(1) != 2, V(1) < V(2), V(1) + 2, 3 + V(1), sum([V(1), V(2)], V(0)) if False else 0)
for f in (lambda: V(1) < 3, lambda: 3 < V(1), lambda: V(1) + "a", lambda: "a" + V(1), lambda: V(1) <= V(2), lambda: V(1) * 2):
    try:
        print(f())
    except TypeError as ex:
        print(ex)
def matrix(a, b):
    return a[..., 0] if False else (a, ..., b)
print(matrix(1, 2), [x for x in (..., None, NotImplemented)], ...  if False else "ok")
