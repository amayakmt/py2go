print(True + True, True * 5, False - 1, True / 2, True // 1, -True, ~True, ~False, True << 3, True & 3, True | 2, True ^ True, True ** 2, abs(False), int(True), float(True), complex(True), str(True), repr(False), True == 1, False == 0.0, True is 1 if False else 0, hash(True), hash(False))
print(isinstance(True, int), isinstance(True, bool), isinstance(1, bool), issubclass(bool, int), type(True).__name__, type(True + 1).__name__, type(True & True).__name__, type(True | False).__name__, type(True ^ True).__name__, type(~True).__name__, type(True and 5).__name__, type(not 5).__name__)
print(sum([True, True, False]), sum(x > 2 for x in range(6)), [True] * 3, [1, 2, 3][True], "abc"[True], (10, 20)[False], {True: "a", 1: "b"}, {1: "x"}[True], len({True, 1, 1.0}), [True, 1, 1.0].count(1), max(True, 0), min(True, 2), sorted([True, 0, 1, False]), True < 2, False >= 0, round(True), divmod(True, 2), pow(True, 3), bin(True), hex(True), format(True, "d"), format(True, ""), f"{True:5}|", f"{True!s}", "%d %s %r" % (True, True, True), int("1") == True)
print(bool(0), bool(0.0), bool(""), bool([]), bool(()), bool({}), bool(None), bool(0j), bool(b""), bool(1), bool(-1), bool(0.1), bool("a"), bool([0]), bool(2 ** 70), bool(-2 ** 70), bool(float("nan")), bool(range(0)), bool(set()), bool(frozenset()), bool(1j), bool(bytearray(1)), bool(...), bool(NotImplemented) if False else 0)
print(True.bit_length(), True.real, True.imag, True.numerator, True.denominator, True.conjugate(), (True).to_bytes(1, "big"), bool.__name__, int(False) + 1, 1 if True else 2, [x for x in (True, False)], all([]), any([]), all([True, 1, "x"]), any([0, "", None, 0.0]), not [], not [0])
for fn in (lambda: True.foo, lambda: bool.__bases__ if False else int.nope):
    try:
        print(fn())
    except (TypeError, AttributeError) as e:
        print(type(e).__name__, e)
