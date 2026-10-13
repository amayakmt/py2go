import_vals = [0, 1, -1, True, False, 0.0, 1.0, -0.0, 2 ** 53, 2 ** 53 + 1, float(2 ** 53), 2 ** 63, float(2 ** 63), 2 ** 64, float(2 ** 64), 0.5, 1e300, 10 ** 300, 10 ** 400, float("inf")]
ops = [("==", lambda a, b: a == b), ("<", lambda a, b: a < b), ("<=", lambda a, b: a <= b), (">", lambda a, b: a > b), ("!=", lambda a, b: a != b)]
for name, op in ops:
    rows = []
    for a in import_vals:
        rows.append("".join("1" if op(a, b) else "0" for b in import_vals))
    print(name, rows)
print([hash(v) == hash(w) for v in import_vals[:8] for w in import_vals[:8] if v == w][:30])
print(1 < 2 < 3, 1 < 3 < 2, 2 ** 70 > 2 ** 69 > 1, 0 <= False < True <= 1, 1 == 1.0 == True, 1 != 2 != 1, 0.5 < 1 > 0.25, 10 ** 400 > 1e308, 10 ** 400 > float("inf") if False else True)
print(max(1, 1.0), max(1.0, 1), min(True, 1), max(True, 1), min(2 ** 70, 2.0 ** 70), max(2 ** 70, 2.0 ** 70), sorted([1, 1.0, True]), sorted([True, 1, 1.0]), [1, 1.0] == [1.0, 1], (1, 2) == (1.0, 2.0), {1: "a"} == {1.0: "a"}, {1, 2} == {2.0, 1.0})
