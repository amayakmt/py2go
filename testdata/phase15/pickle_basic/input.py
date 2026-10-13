import pickle


class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

    def __eq__(self, o):
        return isinstance(o, Point) and self.x == o.x and self.y == o.y


class Empty:
    pass


samples = [
    None, True, False, 0, 1, 255, 256, 65535, 65536, -1, 2**31 - 1, 2**31, -2**31, -2**31 - 1,
    2**100, -2**100, 1.5, -0.0, 3.14159, "", "abc", "héllo\n\\", b"", b"xyz", bytearray(b"ab"),
    (), (1,), (1, 2), (1, 2, 3), (1, 2, 3, 4), [], [1, 2, 3], [[1], [2]], {}, {"a": 1, "b": [1, 2]},
    {1, 2, 3}, frozenset([4, 5]), set(), 1 + 2j, range(3, 10, 2), "x" * 300, list(range(2500)),
]
for proto in range(6):
    for s in samples:
        data = pickle.dumps(s, proto)
        print(proto, data[:60], len(data))
        back = pickle.loads(data)
        assert back == s, (proto, s, back)

shared = [1, 2]
obj = [shared, shared, (shared,)]
for proto in range(6):
    d = pickle.dumps(obj, proto)
    print(proto, d)
    b = pickle.loads(d)
    assert b[0] is b[1] and b == obj

rec = []
rec.append(rec)
for proto in range(6):
    d = pickle.dumps(rec, proto)
    print(d)
    b = pickle.loads(d)
    assert b[0] is b

for proto in range(6):
    p = Point(1, "two")
    d = pickle.dumps(p, proto)
    print(proto, d)
    q = pickle.loads(d)
    assert q == p and type(q) is Point
    print(pickle.dumps(Empty(), proto))
    print(type(pickle.loads(pickle.dumps(Empty(), proto))).__name__)
    print(pickle.dumps(ValueError("bad", 3), proto))
    e = pickle.loads(pickle.dumps(ValueError("bad", 3), proto))
    print(repr(e))
    print(pickle.dumps(int, proto), pickle.loads(pickle.dumps(Point, proto)) is Point)

print(pickle.DEFAULT_PROTOCOL, pickle.HIGHEST_PROTOCOL)
print(pickle.loads(pickle.dumps([Point(1, 2), Point(3, 4)])) == [Point(1, 2), Point(3, 4)])

for bad in [b"", b"\x80\x04", b"?", b"\x80\x09.", b"(", b"N"]:
    try:
        pickle.loads(bad)
    except Exception as e:
        print(type(e).__name__, e)
try:
    pickle.dumps(1, 9)
except ValueError as e:
    print(e)
try:
    pickle.loads("abc")
except TypeError as e:
    print(e)
