r: range = range(2, 20, 3)
print(r, len(r), r[0], r[-1], r[2])
print(list(r), sum(r), min(r), max(r))
print(5 in r, 6 in r, 20 in r, 2 in r)
print(range(5), range(1, 5), range(0, 10, 2), range(5, 0, -1))
print(list(range(5, 0, -2)), list(range(0)), list(range(3, 3)), len(range(10, 0, -3)))
print(range(3) == range(3), range(3) == range(4), range(0) == range(5, 5), range(1, 2) == range(1, 5, 9))
n: int = 4
for i in range(n):
    print(i, end=" ")
print()
squares: list[int] = [i * i for i in range(n)]
print(squares, list(reversed(range(n))), bool(range(0)), bool(range(1)))
print(sorted(range(5, 0, -1)), sum(range(101)))
