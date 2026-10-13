class P:
    def __init__(self, x, y):
        self.x, self.y = x, y
    def __eq__(self, o):
        return isinstance(o, P) and (self.x, self.y) == (o.x, o.y)
    def __hash__(self):
        return hash((self.x, self.y))
    def __repr__(self):
        return f"P({self.x}, {self.y})"
class Q:
    def __init__(self, x):
        self.x = x
    def __eq__(self, o):
        return isinstance(o, Q) and self.x == o.x
class R:
    __hash__ = None
class S:
    def __hash__(self):
        return 42
class T:
    def __hash__(self):
        return -1
class U:
    def __hash__(self):
        return 2 ** 70 + 5
print(P(1, 2) == P(1, 2), hash(P(1, 2)) == hash(P(1, 2)), P(1, 2) in [P(1, 2)], sorted([P(1, 2), P(0, 5)], key=lambda p: p.x), hash(S()), hash(T()), hash(U()) == hash(2 ** 70 + 5), hash(U()))
for fn in (lambda: hash(Q(1)), lambda: hash(R()), lambda: hash([1]), lambda: hash({}), lambda: hash({1}), lambda: hash((1, [2])), lambda: hash(frozenset([1, 2])) == hash(frozenset([2, 1])), lambda: hash(bytearray()), lambda: hash(float("nan")) is not None):
    try:
        print(fn())
    except TypeError as e:
        print(e)
print(hash(1) == 1, hash(-1), hash(-2), hash(0), hash(10), hash(2 ** 61 - 1), hash(2 ** 61), hash(-2 ** 61), hash(2 ** 62 + 3), hash(True), hash(False), hash(None) == hash(None), hash("") == 0, hash(()) == hash(()), hash((1, 2)) != hash((2, 1)), hash("a") == hash("a"), hash(b"a") == hash("a"), hash(1.5) == hash(1.5), hash(3.0) == 3)
