items = [b"pear", b"apple", b"fig", b"apple", b"", b"\xff", b"Zed", b"a\x00"]
print(sorted(items), sorted(items, reverse=True), sorted(items, key=len), sorted(set(items)), items.count(b"apple"), items.index(b"fig"), b"fig" in items, max(items), min(items))
d = {b"x": 1, b"y": [b"a"], b"z": {b"k": b"v"}}
d[b"w"] = d.get(b"w", 0) + 1
print(d, list(d), sorted(d.items()), b"x" in d, d.pop(b"x"), len(d))
s = {b"a", b"b", b"a", bytes([97])}
print(sorted(s), len(s), s == {b"a", b"b"}, frozenset(s) == s, b"a" in frozenset(s))
t = tuple(items[:3])
print(t, t == (b"pear", b"apple", b"fig"), hash(t) == hash((b"pear", b"apple", b"fig")), t + (b"!",), t * 2)
ba = [bytearray(b"b"), bytearray(b"a")]
print(sorted(ba), ba, ba[0] == b"b", b"b" == ba[0], [x.decode() for x in ba])
try:
    {bytearray(b"x"): 1}
except TypeError as e:
    print(e)
try:
    print(bytearray(b"a") < "a")
except TypeError as e:
    print(e)
