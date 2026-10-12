print(list((x, y) for x in range(3) for y in range(x + 1)))
print(sum(x * y for x in range(4) for y in range(4) if x < y))
matrix: list[list[int]] = [[1, 2], [3, 4], [5, 6]]
print(sum(v for row in matrix for v in row))
print(list(v for row in matrix for v in row if v % 2 == 0))
print(any(v == 4 for row in matrix for v in row), all(v < 6 for row in matrix for v in row))
print(list((a, b, c) for a in range(2) for b in range(2) for c in range(2) if a + b + c == 2))
