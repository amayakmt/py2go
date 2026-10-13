print(True + 1, True + True, True * 5, False - 1, -True, ~True, True / 2, True // 1, True % 2, True ** 3, True << 4)
print(True == 1, False == 0, True is True, True < 2, False < True, True >= 1, 1 == 1.0 == True, True != 2)
print(sum([True, True, False]), sum(x > 2 for x in range(10)), max(True, 0), min(True, 2), sorted([True, 0, 2, False]), [True, False].count(1))
print(isinstance(True, int), isinstance(1, bool), isinstance(True, bool), type(True + 1).__name__, type(True & False).__name__, type(True | 2).__name__, True & True, True | False, True ^ True)
d = {True: "t", 1: "one", 2: "two"}
print(d, len(d), d[1], d[True], {0: "z"}[False], [10, 20][True], "ab"[True], (5, 6)[False])
print(bool(1), bool(0.0), bool(""), bool("x"), bool([]), bool([0]), bool(None), bool(2 ** 70), int(True), float(True), str(True), repr(False), True and 5, False or 7)
print(f"{True}", f"{True:d}", f"{True:5}|", f"{False:>6}|", "%d %s %r" % (True, True, True), format(True), format(True, "d"), format(True, "b"))
x = True
x += True
print(x, type(x).__name__)
