for i, b in enumerate(b"abc", 10):
    print(i, b, chr(b))
for i, ch in enumerate(bytearray(b"xy")):
    print(i, ch)
print(list(zip(b"ab", "cd", range(5), frozenset([7]), (1.5, 2.5))), list(zip(*[b"ab", b"cd"])), dict(zip(b"ab", b"cd")), list(enumerate(frozenset(["only"]))), list(enumerate(range(10 ** 20, 10 ** 20 + 2))))
print(list(reversed(b"abc")), list(reversed(bytearray(b"ab"))), list(reversed(range(3))), sorted(b"cab", reverse=True), sorted(frozenset([3, 1, 2])), list(map(hex, b"\x0a\xff")), list(filter(None, b"\x00a\x00b")), sum(b"\x01\x02"), any(b"\x00\x00"), all(b"\x01\x00"), max(b"abc"), min(bytearray(b"abc")), len(list(range(3))))
print([x for x in b"ab" for y in range(2)], {x: y for x, y in zip(b"ab", range(2))}, {x for x in b"aabb"}, tuple(b"ab"), list(iter(b"ab")), next(iter(b"zz")), list(iter(range(3))), [i for i, _ in enumerate(range(3))], any(c == 97 for c in b"abc"), "".join(chr(c) for c in b"hey"))
a, b, *rest = b"abcde"
x, y = frozenset([5, 6]) if False else (1, 2)
print(a, b, rest, x, y)
