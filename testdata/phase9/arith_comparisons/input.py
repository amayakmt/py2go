def cmp(a, b):
    return a == b, a != b, a < b, a <= b, a > b, a >= b

print(cmp(1, 2))
print(cmp(2.0, 2))
print(cmp(True, 1))
print(cmp("a", "b"))
print(cmp([1, 2], [1, 3]))
print(cmp((1, 2), (1, 2)))
print(1 == 1.0, 1 == True, 0 == False, "1" == 1, [1] == (1,), None == 0)
print(1 < 2 < 3, 3 > 2 > 2, 1 < 2 == 2, 0 <= 0.5 < 1)
print(max(1, 2.5), min(3, 2.0), max("a", "b"), max([1, 5, 3]), min((4, 2, 8)))
print(sorted([3, 1.5, True, 2]), sorted(["b", "a", "C"]), sorted([(2, "a"), (1, "z"), (1, "b")]))
