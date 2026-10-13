d = {}
for k in (1, 1.0, True, 2, 2.5, 2 ** 70, float(2 ** 70), 0, 0.0, -0.0, False, 1 + 0j):
    d[k] = d.get(k, 0) + 1
print(d, len(d))
print(hash(1) == hash(1.0) == hash(True), hash(2.5) == hash(2.5), hash(-1) == hash(-2), hash(2 ** 61 - 1), hash(2 ** 61), hash(2 ** 70), hash(-2 ** 70), hash(0.0) == hash(-0.0), hash(float("inf")), hash(float("-inf")), hash(0.5), hash(1e300), hash(3 + 0j) == hash(3))
print(set([1, 1.0, 2, 2.0, True]), {1.0: "a", 1: "b"}, {0.1 + 0.2: 1}[0.30000000000000004], 1 in {1.0}, 2 ** 70 in {float(2 ** 70)}, (1, 2.0) == (1.0, 2), hash((1, 2.0)) == hash((1.0, 2)))
nan = float("nan")
s = {nan, float("nan")}
print(len(s))
print(hash(1.5), hash(-1.5), hash(0.1), hash(3.14159), hash(1e-10), hash(123456789.125))
