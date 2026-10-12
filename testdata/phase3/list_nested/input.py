grid: list[list[int]] = [[1, 2, 3], [4, 5, 6]]
print(grid, len(grid), len(grid[0]))
print(grid[1][2], grid[-1][-1])
grid[0][1] = 20
grid[1].append(7)
print(grid)
rows: list[list[int]] = [[0] * 3] * 2
rows[0][0] = 9
print(rows)
fresh: list[list[int]] = []
for r in range(2):
    fresh.append([0] * 3)
fresh[0][0] = 9
print(fresh)
flat: list[int] = []
for row in grid:
    for v in row:
        flat.append(v)
print(flat)
print(grid[0][::-1], grid[::-1])
