print([(x, y) for x in range(3) for y in range(3)])
print([(x, y) for x in range(4) for y in range(x)])
print([x * y for x in range(1, 4) for y in range(1, 4) if x != y])
matrix: list[list[int]] = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
print([v for row in matrix for v in row])
print([[row[i] for row in matrix] for i in range(3)])
print([[x * y for y in range(3)] for x in range(3)])
print([v for row in matrix for v in row if v % 2 == 0])
print(sum(sum(row) for row in matrix))
