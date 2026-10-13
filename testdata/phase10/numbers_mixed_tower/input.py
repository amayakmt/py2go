vals = [True, 7, 2 ** 70, 2.5, 3 + 1j]
for a in vals:
    print([a + b for b in vals])
    print([a * b for b in vals])
    print([a == b for b in vals])
print(1 == 1.0 == True, 2 ** 70 == float(2 ** 70), 2 ** 70 + 1 == float(2 ** 70), 3 + 0j == 3, True == 1 + 0j)
print(sorted([3, True, 2.5, 2 ** 70, -2 ** 70, 0, False]), max(1, 2.5, True), min(2 ** 70, 1e30), max(2 ** 70, 1e30))
print(sum([1, 2.5, True]), sum([2 ** 70, 0.5]), sum([1j, 1]), 1 + 2 ** 70, 0.5 + 2 ** 70, abs(-2 ** 70) + abs(-1.5) + abs(True))
print([isinstance(v, int) for v in vals], [isinstance(v, float) for v in vals], [isinstance(v, (int, float)) for v in vals], [type(v).__name__ for v in vals])
print(int(2.9), int(-2.9), int(True), float(7), float(2 ** 70), complex(7), int("7") + 1, float("7") + 1, str(7) + "1")
print(hash(7) == hash(7.0) == hash(True * 7 + 0j) == hash(7 + 0j), hash(2.5) == hash(2.5 + 0j), hash(2 ** 70) == hash(float(2 ** 70)))
