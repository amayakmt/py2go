items = [1, -1, 2 ** 64, -2 ** 64, 1.5, -0.0, float("inf"), float("nan"), 1e22, 1e-7, 3j, 1 + 2j, -1j, True, None, b"by\x00tes", bytearray(b"ba"), "s'q", 'd"q', "both'\"", frozenset([1]), frozenset(), (), (1,), (1, 2), [], {}, set(), range(3), slice(1, 2), ..., NotImplemented]
for it in items:
    print(repr(it), str(it), type(it).__name__, bool(it) if not isinstance(it, float) or it == it else "nan-bool")
print(items[:8])
print({k: v for k, v in zip("abcdef", items[:6])})
print([items[2:5], (items[5:8], items[8:10])], {"k": [(1, 2 ** 70)], 5: {2 ** 65: b"x"}}, [[[]]], [(), ((),)], ((1, (2, (3,))),))
print(sorted(i for i in items if isinstance(i, (int, float)) and not isinstance(i, bool) and i == i))
print(f"{items!r}"[:60], str(items)[-30:], len(repr(items)), "%r %s" % (items[3], items[3]), "{} {!r}".format(items[18], items[18]))
