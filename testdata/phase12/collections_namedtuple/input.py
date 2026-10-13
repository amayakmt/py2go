from collections import namedtuple
import collections
Point = namedtuple("Point", "x y")
Person = collections.namedtuple("Person", ["name", "age", "city"], defaults=["nowhere"])
Triple = namedtuple("Triple", "a, b, c")
p = Point(1, 2)
q = Point(y=5, x=4)
x, y = p
print(p, q, x, y, p.x, p[1], p[-1], len(p), list(p), tuple(p), p == (1, 2), p == Point(1, 2), p != q, p < q, max(p), sum(p), 2 in p)
print(p._asdict(), p._replace(x=10), Point._make([7, 8]), Point._fields, Person._fields, p.count(1), p.index(2))
per = Person("ann", 30)
print(per, per.city, per._replace(city="rome").city, Person("bob", 5, "x"), Person._fields, Triple(1, 2, 3))
pts = [Point(3, 4), Point(1, 2), Point(1, 1)]
print(sorted(pts), min(pts), [pt.x for pt in pts], {pt: i for i, pt in enumerate(pts)}[Point(1, 2)], len(set(pts + [Point(1, 2)])), hash(p) == hash((1, 2)), sum(pt.x * pt.y for pt in pts))
def dist(a, b):
    return ((a.x - b.x) ** 2 + (a.y - b.y) ** 2) ** 0.5
print(dist(Point(0, 0), Point(3, 4)), Point(1, 2)._asdict()["y"])
