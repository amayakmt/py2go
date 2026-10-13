class Node:
    def __init__(self, v: int) -> None:
        self.v = v

a = Node(1)
b = Node(1)
c = a
print(a is c, a is b, a == b, a == c, hash(a) == hash(c), hash(a) == hash(b), id(a) == id(c), id(a) == id(b), {a: 1, b: 2, c: 3}[a], len({a, b, c}), [a, b, c].count(a), [a, b].index(b))
d1 = {"x": 1}
d2 = d1
d3 = dict(d1)
print(d1 is d2, d1 is d3, d1 == d3, id(d1) == id(d2), id(d1) != id(d3), hash("x") == hash("x"), hash(1) == hash(1.0), hash((1, "a")) == hash((1, "a")), hash(None) == hash(None), hash(True) == 1)
x: int = 5
y: int = 5
s1: str = "abc"
s2: str = "ab" + "c"
print(x is y, x == y, s1 == s2, id(x) == id(y), None is None, (x is not None), type(x) is int, type(s1) is str, type(a) is Node, type(a) is type(b), isinstance(a, Node))
