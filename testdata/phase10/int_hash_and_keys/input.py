print(hash(0), hash(1), hash(-1), hash(-2), hash(2 ** 61 - 1), hash(2 ** 61), hash(2 ** 62), hash(2 ** 100), hash(-2 ** 100), hash(10 ** 30))
print(hash(2 ** 61 - 1) == hash(float(2 ** 61 - 1)), hash(2 ** 52) == hash(float(2 ** 52)), hash(10 ** 20) == hash(1e20), hash(True) == hash(1))
d = {}
d[10 ** 30] = "big"
d[2 ** 64] = "u64"
d[1] = "one"
d[1.0] = "float one"
d[10 ** 30] = "big again"
print(d, len(d), d[10 ** 30], d[2 ** 64], 10 ** 30 in d, 10 ** 31 in d, d.get(float(10 ** 30)))
s = {2 ** 70, 2 ** 70 + 0, 2 ** 71, 1, 1.0, True, 2 ** 70 - 1 + 1}
print(len(s), sorted(s, key=lambda v: float(v)))
print({10 ** 20: 1} == {10 ** 20: 1}, {2 ** 65: "a"}[2 ** 65], (2 ** 70, "x") in {(2 ** 70, "x")}, {(2 ** 70, 1): 5}[(2 ** 70, 1)])
