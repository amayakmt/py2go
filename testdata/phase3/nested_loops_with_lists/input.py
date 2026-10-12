matrix: list[list[int]] = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
total: int = 0
for row in matrix:
    for v in row:
        if v % 2 == 0:
            continue
        total += v
print(total)
diag: list[int] = []
for i in range(len(matrix)):
    diag.append(matrix[i][i])
print(diag, sum(diag))
for i, row in enumerate(matrix):
    for j, v in enumerate(row):
        if v == 5:
            print("found at", i, j)
            break
found: bool = False
for row in matrix:
    for v in row:
        if v > 7:
            found = True
            break
    if found:
        break
print(found)
